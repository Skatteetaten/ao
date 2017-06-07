#!/bin/bash
#
# Build AOC v5
version=5.0
majorversion=5
minorversion=1

if [ ! -f buildnumber.txt ]; then
  echo "1" >buildnumber.txt
fi

buildnumber=$(<buildnumber.txt)
buildnumber=$(($buildnumber+1))
branch=$(git branch | sed -n -e 's/^\* \(.*\)/\1/p')
echo $buildnumber >buildnumber.txt

go build -ldflags "-X github.com/skatteetaten/aoc/pkg/versionutil.majorVersion=$majorversion -X github.com/skatteetaten/aoc/pkg/versionutil.minorVersion=$minorversion -X github.com/skatteetaten/aoc/pkg/versionutil.branch=$branch -X github.com/skatteetaten/aoc/pkg/versionutil.buildnumber=$buildnumber -X github.com/skatteetaten/aoc/pkg/versionutil.buildstamp=`date '+%Y-%m-%d_%H:%M:%S%p'` -X github.com/skatteetaten/aoc/pkg/versionutil.githash=`git rev-parse HEAD`"
