#!/bin/bash
rm -fr encoding.go
sszgen --path . -objs SSZFile_v1 -output encoding.go -include big/
