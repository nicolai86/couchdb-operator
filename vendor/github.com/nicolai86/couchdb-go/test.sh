#!/usr/bin/env bash
#
set -e

couchdb_id=$(docker run -p 5985:5984 -d couchdb:1.6.1)
function cleanup {
  docker stop $couchdb_id > /dev/null
  docker rm $couchdb_id > /dev/null
}
trap cleanup EXIT

echo Waiting for couchdb to start…
while ! curl http://localhost:5985 -s -q > /dev/null; do sleep 1; done

export COUCHDB_HOST_PORT=http://localhost:5985
echo Running tests…
go test ./... -v
