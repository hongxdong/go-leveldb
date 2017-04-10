// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
  "testing"
)

const kCacheSize = 1000
var current_ Cache = NewLRUCache(kCacheSize)

func TestCache_HitAndMiss(t *testing.T) {
}

