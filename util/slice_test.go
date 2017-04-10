// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"testing"
)

func TestSlice(t *testing.T) {
  var s = NewSlice([]byte("HelloWorld"))

  if s.size() != 10 {
    t.Fatalf("Size error")
  }

  if s.empty() {
    t.Fatalf("Empty error")
  }

  if s.at(0) != 'H' {
    t.Fatalf("at error")
  }

  var b = NewSlice([]byte("WellHelloMac"))
  b.remove_prefix(4)

  if string(b.data()) != "HelloMac" {
    t.Fatalf("remove_prefix error")
  }

  if b.ToString() != "HelloMac" {
    t.Fatalf("remove_prefix error")
  }

  if b.size() != 8 {
    t.Fatalf("remove_prefix error")
  }

  if s.compare(b) <= 0 {
    t.Fatalf("compare error")
  }

  var c = NewSlice([]byte("Hello"))

  if !s.starts_with(c) {
    t.Fatalf("starts_with error")
  }

  if s.starts_with(b) {
    t.Fatalf("starts_with error")
  }

  if s.Equal(b) {
    t.Fatalf("Equal error")
  }

  if !s.NotEqual(b) {
    t.Fatalf("NotEqual error")
  }

  var e = NewSlice([]byte(""))

  if !e.empty() {
    t.Fatalf("NotEqual error")
  }
}

