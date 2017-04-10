// Copyright (c) 2016 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
  "testing"
)

func TestCRC32_StandardResults(t *testing.T) {
  // From rfc3720 section B.4. StandardResults.

  var buf = make([]byte, 32)

  buf[0] = 0
  if (NewCRC32(buf) != 0x8a9136aa) {
    t.Fatalf("CRC32 error.")
  }

  for i := 0; i < len(buf); i++ {
    buf[i] = 0xff
  }
  if (NewCRC32(buf) != 0x62a8ab43) {
    t.Fatalf("CRC32 error.%#x", NewCRC32(buf))
  }

  for i := 0; i < 32; i++ {
    buf[i] = byte(i)
  }
  if (NewCRC32(buf) != 0x46dd794e) {
    t.Fatalf("CRC32 error.")
  }

  for i := 0; i < 32; i++ {
    buf[i] = byte(31 - i);
  }
  if (NewCRC32(buf) != 0x113fdb5c) {
    t.Fatalf("CRC32 error.")
  }

  data := []byte {
    0x01, 0xc0, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00,
    0x14, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x04, 0x00,
    0x00, 0x00, 0x00, 0x14,
    0x00, 0x00, 0x00, 0x18,
    0x28, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00,
    0x02, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00,
  }
  if (NewCRC32(data) != 0xd9963a56) {
    t.Fatalf("CRC32 error.")
  }
}

func TestCRC32_NewValue(t *testing.T) {
  if NewCRC32([]byte("a")) == NewCRC32([]byte("foo")) {
    t.Fatalf("CRC32 error.")
  }
}

func TestCRC32_ExtendCRC32(t *testing.T) {
  a := NewCRC32([]byte("hello world"))
  b := NewCRC32([]byte("hello ")).ExtendCRC32([]byte("world"))
  if a != b {
    t.Fatalf("CRC32 error.")
  }
}

func TestCRC32_Mask(t *testing.T) {
  crc := NewCRC32([]byte("foo")).Value()
  if crc == MaskCRC32(crc) {
    t.Fatalf("CRC32 error.")
  }

  if crc == MaskCRC32(MaskCRC32(crc)) {
    t.Fatalf("CRC32 error.")
  }

  if crc != UnmaskCRC32(MaskCRC32(crc)) {
    t.Fatalf("CRC32 error.")
  }

  if crc != UnmaskCRC32(UnmaskCRC32(MaskCRC32(MaskCRC32(crc)))) {
    t.Fatalf("CRC32 error.")
  }
}
