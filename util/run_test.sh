#!/bin/bash

echo "test cache"
go test cache_test.go cache.go slice.go hash.go

echo "test crc32c"
go test crc32c_test.go crc32c.go

echo "test slice"
go test slice_test.go slice.go

echo "test hash"
go test hash_test.go hash.go

