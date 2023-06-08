#!/bin/bash
pwd
go build -buildvcs=false
mv rocketpool-cli ~/bin/rocketpool
chmod u+x ~/bin/rocketpool
