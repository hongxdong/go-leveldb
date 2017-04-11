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

func Deleter(key *Slice, v interface{}) {
  current_.deleted_keys_   = append(current_.deleted_keys_, DecodeKey(key))
  current_.deleted_values_ = append(current_.deleted_values_, DecodeValue(v))
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

func TestCache_HitAndMiss(t *testing.T) {
  fmt.Println("Run TestCache")

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

  ASSERT_EQ(1, len(current_.deleted_keys_))
  ASSERT_EQ(100, current_.deleted_keys_[0])
  ASSERT_EQ(101, current_.deleted_values_[0])
  // fmt.Printf("(%v, %T)\n", current_.deleted_values_, current_.deleted_values_)
}


