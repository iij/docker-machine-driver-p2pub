#! /bin/sh

set -x
cd $(dirname $0)/bin
go build -o docker-machine-driver-p2pub
