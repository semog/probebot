#!/bin/bash

# Turn off dynamic linking of glibc. Use the pure Go implementation.
# Strip the debug symbols out of the executable.
go build -tags netgo -ldflags '-s'
