#!/usr/bin/env bash

docker run --rm --privileged \
    -v $(pwd):/go/github.com/VolantMQ/vlapi \
    -w /go/github.com/VolantMQ/vlapi \
    mailchain/goreleaser-xcgo goreleaser --snapshot --rm-dist
