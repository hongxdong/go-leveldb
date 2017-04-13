// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
  "strconv"
)

// a == b
func ASSERT_EQ(a int, b int) {
  if (a != b) {
    var s string = "a:" + strconv.Itoa(a) + " b:" + strconv.Itoa(b)
    panic("ASSERT_EQ() error. " + s)
  }
}

// a <= b
func ASSERT_LE(a int, b int) {
  if (a > b) {
    var s string = " a:" + strconv.Itoa(a) + " b:" + strconv.Itoa(b)
    panic("ASSERT_LE() error. " + s)
  }
}

// a != b
func ASSERT_NE(a uint64, b uint64) {
  if (a == b) {
    var s string = " a:" + strconv.FormatUint(a, 10) + " b:" + strconv.FormatUint(b, 10)
    panic("ASSERT_NE() error. " + s)
  }
}
