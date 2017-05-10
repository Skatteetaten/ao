#!/bin/bash
#
# Build AOC v5
version=5.0.0
buildnumber=$(<buildnumber.txt)
buildnumber=$(($buildnumber+1))
echo $buildnumber >buildnumber.txt

go build -ldflags "-X github.com/skatteetaten/aoc/cmd.buildnumber=$buildnumber -X github.com/skatteetaten/aoc/cmd.version=$version -X github.com/skatteetaten/aoc/cmd.buildstamp=`date '+%Y-%m-%d_%H:%M:%S%p'` -X github.com/skatteetaten/aoc/cmd.githash=`git rev-parse HEAD`"
