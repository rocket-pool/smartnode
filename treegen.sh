#!/bin/sh

docker run --rm -v $PWD/out:/out rocketpool/treegen /treegen -o /out "$@"
