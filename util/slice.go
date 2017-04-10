// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Slice is a simple structure containing a pointer into some external
// storage and a size.  The user of a Slice must ensure that the slice
// is not used after the corresponding external storage has been
// deallocated.
//
// Multiple threads can invoke const methods on a Slice without
// external synchronization, but if any of the threads may call a
// non-const method, all threads accessing the same Slice must use
// external synchronization.

package util

import (
  "bytes"
)

type Slice struct {
  data_ []byte
  size_ uint64
}

// Create a slice
func NewSlice(data []byte) *Slice {
  return &Slice{data, uint64(len(data))}
}

// Return data
func (s *Slice) data() []byte {
  return s.data_
}

// Return the length (in bytes) of the referenced data
func (s *Slice) size() uint64 {
  return s.size_
}

// Return true iff the length of the referenced data is zero
func (s *Slice) empty() bool {
  return s.size_ == 0
}

// Return the ith byte in the referenced data.
// REQUIRES: n < size()
func (s *Slice) at(n uint64) byte {
  if (n >= s.size()) {
    panic("Slice at() error")
  }
  return s.data_[n]
}

// Change this slice to refer to an empty array
func (s *Slice) clear() {
  s.data_ = nil
  s.size_ = 0
}

// Drop the first "n" bytes from this slice.
func (s *Slice) remove_prefix(n uint64) {
  if (n > s.size()) {
    panic("Slice remove_prefix() error")
  }
  s.data_ = s.data_[n:]
  s.size_ -= n
}

// Return a string that contains the copy of the referenced data.
func (s *Slice) ToString() string {
  return string(s.data_)
}

// Three-way comparison.  Returns value:
//   <  0 iff "*this" <  "b",
//   == 0 iff "*this" == "b",
//   >  0 iff "*this" >  "b"
func (s *Slice) compare(b *Slice) int {
  return bytes.Compare(s.data_, b.data_)
}

// Return true iff "x" is a prefix of "*this"
func (s *Slice) starts_with(x *Slice) bool {
  return bytes.HasPrefix(s.data_, x.data_)
}

func (s *Slice) Equal(b *Slice) bool {
  return bytes.Equal(s.data_, b.data_)
}

func (s *Slice) NotEqual(b *Slice) bool {
  return !s.Equal(b)
}

