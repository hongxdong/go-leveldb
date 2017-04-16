// Copyright (c) 2017 Hong Xiaodong. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

// C++ use class constructor and destructor
// to automatically Lock and Unlock a mutex.
// But in go, We can use sync.Mutex as following:
// var mutex sync.Mutex
// mutex.Lock()
// defer mutex.Unlock()
