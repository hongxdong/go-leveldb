// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
  "hash/crc32"
)

// CRC32 using Castagnoli's polynomial, CRC32C is an alias to Castagnoli
// Using 32bit CRC result
// Note: We import and use "hash/crc32" package for crc32 calculation,
//       Intel CRC32 acceleration is enabled by default.
type CRC uint32

// cheat table of CRC32
var kCheatTable = crc32.MakeTable(crc32.Castagnoli)


// Create new CRC32 with data bytes
func NewCRC32(data []byte) CRC {
  return CRC(0).ExtendCRC32(data)
}

// Extend CRC with data bytes
func (i CRC) ExtendCRC32(data []byte) (CRC) {
  return CRC(crc32.Update(uint32(i), kCheatTable, data))
}

// return the value
func (i CRC) Value() uint32 {
  return uint32(i)
}

// Return a masked representation of crc.
//
// Motivation: it is problematic to compute the CRC of a string that
// contains embedded CRCs.  Therefore we recommend that CRCs stored
// somewhere (e.g., in files) should be masked before being stored.
const kMaskDelta = 0xa282ead8

func MaskCRC32(crc uint32) uint32 {
  return uint32((crc >> 15) | (crc << 17)) + kMaskDelta
}

// Return the crc whose masked representation is masked_crc.
func UnmaskCRC32(masked_crc uint32) uint32 {
  var rot = masked_crc - kMaskDelta
  return ((rot >> 17) | (rot << 15))
}
