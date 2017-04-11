// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

func ASSERT_EQ(a int, b int) {
  if (a != b) {
    panic("ASSERT_EQ() error")
  }
}

