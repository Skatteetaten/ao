#!/bin/bash
#
# Build AOC v5
version=5.0
buildnumber=$(<buildnumber.txt)
buildnumber=$(($buildnumber+1))
branch=$(git branch | sed -n -e 's/^\* \(.*\)/\1/p')
echo $buildnumber >buildnumber.txt

go build -ldflags "-X github.com/skatteetaten/aoc/cmd.branch=$branch -X github.com/skatteetaten/aoc/cmd.buildnumber=$buildnumber -X github.com/skatteetaten/aoc/cmd.version=$version -X github.com/skatteetaten/aoc/cmd.buildstamp=`date '+%Y-%m-%d_%H:%M:%S%p'` -X github.com/skatteetaten/aoc/cmd.githash=`git rev-parse HEAD`"
