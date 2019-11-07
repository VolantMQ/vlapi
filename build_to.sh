#!/usr/bin/env bash

#export GO111MODULE=off
export GOPROXY=/Users/amr/go/pkg/mod

go build -installsuffix=dynlink -gcflags=-dynlink -buildmode=plugin -o $1/debug.so ./vlplugin/debug
go build -installsuffix=dynlink -gcflags=-dynlink -buildmode=plugin -o $1/health.so ./vlplugin/health
go build -installsuffix=dynlink -gcflags=-dynlink -buildmode=plugin -o $1/persistence_bbolt.so ./vlplugin/vlpersistence/bbolt/plugin
