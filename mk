#!/bin/bash

# Turn off dynamic linking of glibc. Use the pure Go implementation.
go build -tags netgo
