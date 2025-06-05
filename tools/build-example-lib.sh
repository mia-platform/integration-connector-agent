#! /usr/bin/env bash

go build -o processor-lib.so -buildmode=plugin processor.go
