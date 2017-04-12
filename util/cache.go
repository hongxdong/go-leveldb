// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// A Cache is an interface that maps keys to values.  It has internal
// synchronization and may be safely accessed concurrently from
// multiple threads.  It may automatically evict entries to make room
// for new entries.  Values have a specified charge against the cache
// capacity.  For example, a cache where the values are variable
// length strings, may use the length of the string as the charge for
// the string.
//
// A builtin cache implementation with a least-recently-used eviction
// policy is provided.  Clients may use their own implementations if
// they want something more sophisticated (like scan-resistance, a
// custom eviction policy, variable cache sizing, etc.)

package util

import (
  "sync"
  //"fmt"
)

// Create a new cache with a fixed size capacity.  This implementation
// of Cache uses a least-recently-used eviction policy.
func NewLRUCache(capacity uint64) Cache {
  return ConstructShardedLRUCache(capacity)
}

// Opaque handle to an entry stored in the cache.
type CacheHandle interface{}

type Cache interface {
  // Insert a mapping from key->value into the cache and assign it
  // the specified charge against the total cache capacity.
  //
  // Returns a handle that corresponds to the mapping.  The caller
  // must call this->Release(handle) when the returned mapping is no
  // longer needed.
  //
  // When the inserted entry is no longer needed, the key and
  // value will be passed to "deleter".
  Insert(key *Slice, value interface{}, charge uint64, deleter LRUHandleDeleter) CacheHandle

  // If the cache has no mapping for "key", returns NULL.
  //
  // Else return a handle that corresponds to the mapping.  The caller
  // must call this->Release(handle) when the returned mapping is no
  // longer needed.
  Lookup(key *Slice) CacheHandle

  // Release a mapping returned by a previous Lookup().
  // REQUIRES: handle must not have been released yet.
  // REQUIRES: handle must have been returned by a method on *this.
  Release(handle CacheHandle)

  // Return the value encapsulated in a handle returned by a
  // successful Lookup().
  // REQUIRES: handle must not have been released yet.
  // REQUIRES: handle must have been returned by a method on *this.
  Value(handle CacheHandle) interface{}

  // If the cache contains entry for key, erase it.  Note that the
  // underlying entry will be kept around until all existing handles
  // to it have been released.
  Erase(key *Slice)

  // Return a new numeric id.  May be used by multiple clients who are
  // sharing the same cache to partition the key space.  Typically the
  // client will allocate a new id at startup and prepend the id to
  // its cache keys.
  NewId() uint64

  // Remove all cache entries that are not actively in use.  Memory-constrained
  // applications may wish to call this method to reduce memory usage.
  // Default implementation of Prune() does nothing.  Subclasses are strongly
  // encouraged to override the default implementation.  A future release of
  // leveldb may change Prune() to a pure abstract method.
  Prune()

  // Return an estimate of the combined charges of all elements stored in the
  // cache.
  TotalCharge() uint64

  // LRU_Remove(e *CacheHandle)
  // LRU_Append(e *CacheHandle)
  // Unref(e *CacheHandle)
}

// LRU cache implementation
//
// Cache entries have an "in_cache" boolean indicating whether the cache has a
// reference on the entry.  The only ways that this can become false without the
// entry being passed to its "deleter" are via Erase(), via Insert() when
// an element with a duplicate key is inserted, or on destruction of the cache.
//
// The cache keeps two linked lists of items in the cache.  All items in the
// cache are in one list or the other, and never both.  Items still referenced
// by clients but erased from the cache are in neither list.  The lists are:
// - in-use:  contains the items currently referenced by clients, in no
//   particular order.  (This list is used for invariant checking.  If we
//   removed the check, elements that would otherwise be on this list could be
//   left as disconnected singleton lists.)
// - LRU:  contains the items not currently referenced by clients, in LRU order
// Elements are moved between these lists by the Ref() and Unref() methods,
// when they detect an element in the cache acquiring or losing its only
// external reference.

// An entry is a variable length heap-allocated structure.  Entries
// are kept in a circular doubly linked list ordered by access time.

type LRUHandleDeleter func(*Slice, interface{})

type LRUHandle struct {
  value      interface{}
  deleter    LRUHandleDeleter
  next_hash  *LRUHandle
  next       *LRUHandle
  prev       *LRUHandle
  charge     uint64      // TODO(opt): Only allow uint32_t?
  key_length uint64
  in_cache   bool        // Whether entry is in the cache.
  refs       uint32      // References, including cache reference, if present.
  hash       uint32      // Hash of key(); used for fast sharding and comparisons
  key_data   []byte      // Beginning of key
}


func (lh *LRUHandle) key() *Slice {
  // For cheaper lookups, we allow a temporary Handle object
  // to store a pointer to a key in "value".
  if (lh.next == lh) {
    return lh.value.(*Slice)
  } else {
    return NewSlice(lh.key_data)
  }
}


// We provide our own simple hash table since it removes a whole bunch
// of porting hacks and is also faster than some of the built-in hash
// table implementations in some of the compiler/runtime combinations
// we have tested.  E.g., readrandom speeds up by ~5% over the g++
// 4.4.3's builtin hashtable.

type HandleTable struct {
  // The table consists of an array of buckets where each bucket is
  // a linked list of cache entries that hash into the bucket.
  length_ uint32
  elems_  uint32
  list_   []*LRUHandle
}

func ConstructHandleTable() HandleTable {
  var ret HandleTable
  ret.Resize()
  return ret
}

func (s *HandleTable) Lookup(key *Slice, hash uint32) *LRUHandle {
  return *s.FindPointer(key, hash)
}

func (s *HandleTable) Insert(h *LRUHandle) *LRUHandle {
  var ptr **LRUHandle = s.FindPointer(h.key(), h.hash)
  var old *LRUHandle = *ptr
  if old == nil {
    h.next_hash = nil
  } else {
    h.next_hash = old.next_hash
  }
  *ptr = h
  if old == nil {
    s.elems_++
    if s.elems_ > s.length_ {
      // Since each cache entry is fairly large, we aim for a small
      // average linked list length (<= 1).
      s.Resize()
    }
  }
  return old
}

func (s *HandleTable) Remove(key *Slice, hash uint32) *LRUHandle {
  var ptr **LRUHandle = s.FindPointer(key, hash)
  var result *LRUHandle = *ptr
  if result != nil {
    *ptr = result.next_hash
    s.elems_--
  }
  return result
}

// Return a pointer to slot that points to a cache entry that
// matches key/hash.  If there is no such cache entry, return a
// pointer to the trailing slot in the corresponding linked list.
func (s *HandleTable) FindPointer(key *Slice, hash uint32) **LRUHandle {
  var ptr **LRUHandle = &s.list_[hash & (s.length_ - 1)]
  for (*ptr != nil) && ((*ptr).hash != hash || key.NotEqual((*ptr).key())) {
    ptr = &(*ptr).next_hash
  }
  return ptr
}

func (s *HandleTable) Resize() {
  var new_length = uint32(4)
  for new_length < s.elems_ {
    new_length *= 2
  }
  new_list := make([]*LRUHandle, new_length)
  var count uint32
  for i := uint32(0); i < s.length_; i++ {
    var h *LRUHandle = s.list_[i]
    for h != nil {
      var next *LRUHandle = h.next_hash
      var hash uint32 = h.hash
      var ptr **LRUHandle = &new_list[hash & (new_length - 1)]
      h.next_hash = *ptr
      *ptr = h
      h = next
      count++
    }
  }
  if (s.elems_ != count) {
    panic("HandleTable Resize() error")
  }
  s.list_ = new_list
  s.length_ = new_length
}

// A single shard of sharded cache.
type LRUCache struct {
  capacity_ uint64      // Initialized before use.
  mutex_    sync.Mutex  // mutex_ protects the following state.
  usage_    uint64

  // Dummy head of LRU list.
  // lru.prev is newest entry, lru.next is oldest entry.
  // Entries have refs==1 and in_cache==true.
  lru_      LRUHandle  // circular doubly linked list ordered by access time.

  // Dummy head of in-use list.
  // Entries are in use by clients, and have refs >= 2 and in_cache==true.
  in_use_   LRUHandle
  table_    HandleTable
}

func ConstructLRUCache() *LRUCache {
  // Make empty circular linked lists.
  var ret = new(LRUCache)
  ret.usage_ = 0
  ret.lru_.next = &ret.lru_
  ret.lru_.prev = &ret.lru_
  ret.in_use_.next = &ret.in_use_
  ret.in_use_.prev = &ret.in_use_
  ret.table_ = ConstructHandleTable()
  return ret
}

func (s *LRUCache) DestructLRUCache() {
  if (s.in_use_.next != &s.in_use_) {   // Error if caller has an unreleased handle
    panic("DestructLRUCache() error")
  }

  for e := s.lru_.next; e != &s.lru_; {
    var next *LRUHandle = e.next
    if !e.in_cache {
      panic("DestructLRUCache() error")
    }
    e.in_cache = false
    if e.refs != 1 {    // Invariant of lru_ list.
      panic("DestructLRUCache() error")
    }
    s.Unref(e)
    e = next
  }
}

func (s *LRUCache) SetCapacity(capacity uint64) {
  s.capacity_ = capacity
}

func (s *LRUCache) Ref(e *LRUHandle) {
  if e.refs == 1 && e.in_cache {    // If on lru_ list, move to in_use_ list.
    s.LRU_Remove(e)
    s.LRU_Append(&s.in_use_, e)
  }
  e.refs++
}

func (s *LRUCache) Unref(e *LRUHandle) {
  if e.refs <= 0 {
    panic("Unref() error")
  }
  e.refs--
  if e.refs == 0 {  // Deallocate.
    if e.in_cache {
      panic("Unref() error")
    }
    e.deleter(e.key(), e.value)
    // fmt.Printf("deleter(%v, %T)\n", e, e)
    // free(e);
  } else if e.in_cache && e.refs == 1 {   // No longer in use; move to lru_ list.
    // fmt.Printf("lru_(%v, %T)\n", e, e)
    s.LRU_Remove(e)
    s.LRU_Append(&s.lru_, e)
  }
}

func (s *LRUCache) LRU_Remove(e *LRUHandle) {
  e.next.prev = e.prev
  e.prev.next = e.next
}

func (s *LRUCache) LRU_Append(list *LRUHandle, e *LRUHandle) {
  // Make "e" newest entry by inserting just before *list
  e.next = list
  e.prev = list.prev
  e.prev.next = e
  e.next.prev = e
}

func (s *LRUCache) Lookup(key *Slice, hash uint32) CacheHandle {
  s.mutex_.Lock()
  var e *LRUHandle = s.table_.Lookup(key, hash)
  if e != nil {
    s.Ref(e)
  }
  s.mutex_.Unlock()
  return e
}

func (s *LRUCache) Release(handle CacheHandle) {
  s.mutex_.Lock()
  s.Unref(handle.(*LRUHandle))
  s.mutex_.Unlock()
}

func (s *LRUCache) Insert(key *Slice, hash uint32, value interface{},
                          charge uint64, deleter LRUHandleDeleter) CacheHandle {
  s.mutex_.Lock()

  var e *LRUHandle = new(LRUHandle)
  e.value = value
  e.deleter = deleter
  e.charge = charge
  e.key_length = key.size()
  e.hash = hash
  e.in_cache = false
  e.refs = 1  // for the returned handle.
  e.key_data = append(e.key_data, key.data() ...)

  if s.capacity_ > 0 {
    e.refs++  // for the cache's reference.
    e.in_cache = true
    s.LRU_Append(&s.in_use_, e)
    s.usage_ += charge
    s.FinishErase(s.table_.Insert(e))
  } // else don't cache.  (Tests use capacity_==0 to turn off caching.)

  for s.usage_ > s.capacity_ && s.lru_.next != &s.lru_ {
    var old *LRUHandle = s.lru_.next
    if old.refs != 1 {
      panic("Insert() error")
    }
    var erased bool = s.FinishErase(s.table_.Remove(old.key(), old.hash))
    if !erased {
      panic("Insert() error")
    }
  }

  s.mutex_.Unlock()
  return e
}

// If e != NULL, finish removing *e from the cache; it has already been removed
// from the hash table.  Return whether e != NULL.  Requires mutex_ held.
func (s *LRUCache) FinishErase(e *LRUHandle) bool {
  if e != nil {
    if !e.in_cache {
      panic("FinishErase() error")
    }
    s.LRU_Remove(e)
    e.in_cache = false
    s.usage_ -= e.charge
    s.Unref(e)
  }
  return e != nil
}

func (s *LRUCache) Erase(key *Slice, hash uint32) {
  s.mutex_.Lock()
  s.FinishErase(s.table_.Remove(key, hash))
  s.mutex_.Unlock()
}

func (s *LRUCache) Prune() {
  s.mutex_.Lock()
  for s.lru_.next != &s.lru_ {
    var e *LRUHandle = s.lru_.next
    if e.refs != 1 {
      panic("Prune() error")
    }
    var erased bool = s.FinishErase(s.table_.Remove(e.key(), e.hash))
    if !erased {  // to avoid unused variable when compiled NDEBUG
      panic("Prune() error")
    }
  }
  s.mutex_.Unlock()
}

func (s *LRUCache) TotalCharge() uint64 {
  s.mutex_.Lock()
  var ret = s.usage_
  s.mutex_.Unlock()
  return ret
}

const kNumShardBits = uint32(4)
const kNumShards    = 1 << kNumShardBits

type ShardedLRUCache struct {
  shard_    [kNumShards]*LRUCache
  id_mutex_ sync.Mutex
  last_id_  uint64
}

func (t *ShardedLRUCache) HashSlice(s *Slice) uint32 {
  return Hash(s.data(), 0)
}

func (t *ShardedLRUCache) Shard(hash uint32) uint32 {
  return hash >> (32 - kNumShardBits)
}

func ConstructShardedLRUCache(capacity uint64) *ShardedLRUCache {
  var slru *ShardedLRUCache = new(ShardedLRUCache)
  slru.last_id_ = 0
  var per_shard uint64 = uint64((capacity + (kNumShards - 1)) / kNumShards)
  for s := 0; s < kNumShards; s++ {
    var lru_cache *LRUCache = ConstructLRUCache()
    slru.shard_[s] = lru_cache
    slru.shard_[s].SetCapacity(per_shard)
  }
  return slru
}

func (t *ShardedLRUCache) Insert(key *Slice, value interface{}, charge uint64, deleter LRUHandleDeleter) CacheHandle {
  var hash uint32 = t.HashSlice(key)
  return t.shard_[t.Shard(hash)].Insert(key, hash, value, charge, deleter)
}

func (t *ShardedLRUCache) Lookup(key *Slice) CacheHandle {
  var hash uint32 = t.HashSlice(key)
  return t.shard_[t.Shard(hash)].Lookup(key, hash)
}

func (t *ShardedLRUCache) Release(handle CacheHandle) {
  var h *LRUHandle = (handle).(*LRUHandle)
  t.shard_[t.Shard(h.hash)].Release(handle)
}

func (t *ShardedLRUCache) Erase(key *Slice) {
  var hash uint32 = t.HashSlice(key)
  t.shard_[t.Shard(hash)].Erase(key, hash)
}

func (t *ShardedLRUCache) Value(handle CacheHandle) interface{} {
  var h *LRUHandle = (handle).(*LRUHandle)
  return h.value
}

func (t *ShardedLRUCache) NewId() uint64 {
  t.id_mutex_.Lock()
  t.last_id_++
  var ret = t.last_id_
  t.id_mutex_.Unlock()
  return ret
}

func (t *ShardedLRUCache) Prune() {
  for s := 0; s < kNumShards; s++ {
    t.shard_[s].Prune()
  }
}

func (t *ShardedLRUCache) TotalCharge() uint64 {
  var total uint64 = 0
  for s := 0; s < kNumShards; s++ {
    total += t.shard_[s].TotalCharge();
  }
  return total
}




