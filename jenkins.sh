#!/bin/bash

type go-junit-report 2> /dev/null || go get -d -u github.com/jstemmer/go-junit-report && go install github.com/jstemmer/go-junit-report
type gocov 2> /dev/null || go get -d github.com/axw/gocov/gocov
type gocov-xml 2> /dev/null || go get -d github.com/AlekSi/gocov-xml

export JUNIT_REPORT=TEST-junit.xml
export COBERTURA_REPORT=coverage.xml

# Go get is not the best way of installing.... :/
export PATH=$PATH:$HOME/go/bin

go mod tidy

make clean build test
