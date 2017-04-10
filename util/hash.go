// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
  "encoding/binary"
)

func Hash(data []byte, seed uint32) uint32 {
  // Similar to murmur hash
  const m = uint32(0xc6a4a793)
  const r = uint32(24)

  var datalen = uint32(len(data))
  var h = seed ^ uint32(datalen * m)
  var i = uint32(0)

  // Pick up four bytes at a time
  for i + 4 <= datalen {
    w := binary.LittleEndian.Uint32(data[i:])
    i += 4
    h += w
    h *= m
    h ^= (h >> 16)
  }

  // Pick up remaining bytes
  switch datalen - i {
  case 3:
    h += uint32(data[i+2]) << 16
    fallthrough
  case 2:
    h += uint32(data[i+1]) << 8
    fallthrough
  case 1:
    h += uint32(data[0])
    h *= m
    h ^= (h >> r)
  }

  return h
}
