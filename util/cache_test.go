// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
  "testing"
  "encoding/binary"
  "fmt"
)

func EncodeKey(k int) []byte {
  var result []byte = make([]byte, 4)
  binary.LittleEndian.PutUint32(result, uint32(k))
  return result
}

func DecodeKey(k *Slice) int {
  if k.size() != 4 {
    panic("DecodeKey() error")
  }
  return int(binary.LittleEndian.Uint32(k.data()))
}

func DecodeValue(v interface{}) int {
  return v.(int)
}

const kCacheSize = 1000

var current_ *CacheTest = ConstructCacheTest()

type CacheTest struct {
  deleted_keys_   []int
  deleted_values_ []int
  cache_   Cache
}

var current_deleted_keys   []int
var current_deleted_values []int

func Deleter(key *Slice, v interface{}) {
  current_deleted_keys   = append(current_deleted_keys, DecodeKey(key))
  current_deleted_values = append(current_deleted_values, DecodeValue(v))
}

func ConstructCacheTest() *CacheTest {
  var cache_test *CacheTest = new(CacheTest)
  cache_test.cache_ = NewLRUCache(kCacheSize)
  return cache_test
}

func (s *CacheTest) Lookup(key int) int {
  var handle CacheHandle = s.cache_.Lookup(NewSlice(EncodeKey(key)))
  var lru_handle *LRUHandle = handle.(*LRUHandle)
  var r int
  if lru_handle == nil {
    r = -1
  } else {
    r = DecodeValue(s.cache_.Value(handle))
  }
  if lru_handle != nil {
    s.cache_.Release(handle)
  }
  return r
}

func (s *CacheTest) Insert(key int, value int, charge uint64) {
  s.cache_.Release(s.cache_.Insert(NewSlice(EncodeKey(key)), value, charge, Deleter))
}

func (s *CacheTest) Erase(key int) {
  s.cache_.Erase(NewSlice(EncodeKey(key)))
}

func TestCache_HitAndMiss(t *testing.T) {
  fmt.Println("Run TestCache_HitAndMiss()")
  current_deleted_keys   = current_deleted_keys[:0]
  current_deleted_values = current_deleted_values[:0]

  ASSERT_EQ(-1, current_.Lookup(100))

  current_.Insert(100, 101, 1)
  ASSERT_EQ(101, current_.Lookup(100))
  ASSERT_EQ(-1, current_.Lookup(200))
  ASSERT_EQ(-1, current_.Lookup(300))

  current_.Insert(200, 201, 1)
  ASSERT_EQ(101, current_.Lookup(100))
  ASSERT_EQ(201, current_.Lookup(200))
  ASSERT_EQ(-1, current_.Lookup(300))

  current_.Insert(100, 102, 1)
  ASSERT_EQ(102, current_.Lookup(100))
  ASSERT_EQ(201, current_.Lookup(200))
  ASSERT_EQ(-1, current_.Lookup(300))

  ASSERT_EQ(1, len(current_deleted_keys))
  ASSERT_EQ(100, current_deleted_keys[0])
  ASSERT_EQ(101, current_deleted_values[0])
  // fmt.Printf("(%v, %T)\n", current_.deleted_values_, current_.deleted_values_)
}

var current_2 *CacheTest = ConstructCacheTest()

func TestCache_Erase(t *testing.T) {
  current_deleted_keys   = current_deleted_keys[:0]
  current_deleted_values = current_deleted_values[:0]

  current_2.Erase(200)
  ASSERT_EQ(0, len(current_2.deleted_keys_))

  current_2.Insert(100, 101, 1)
  current_2.Insert(200, 201, 1)
  current_2.Erase(100)
  // fmt.Printf("(%v, %T)\n", current_deleted_keys, current_deleted_keys)
  ASSERT_EQ(-1,  current_2.Lookup(100))
  ASSERT_EQ(201, current_2.Lookup(200))
  ASSERT_EQ(1,   len(current_deleted_keys))
  ASSERT_EQ(100, current_deleted_keys[0])
  ASSERT_EQ(101, current_deleted_values[0])

  current_2.Erase(100)
  ASSERT_EQ(-1,  current_2.Lookup(100))
  ASSERT_EQ(201, current_2.Lookup(200))
  ASSERT_EQ(1,   len(current_deleted_keys))
}









