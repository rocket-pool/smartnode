#!/bin/bash
rm -fr encoding.go
sszgen --path . -objs SSZFile_v1,SSZFile_v2 -output encoding.go -include big/
