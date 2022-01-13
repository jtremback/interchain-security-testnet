#!/bin/bash
set -eux

# the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# clean up the directory where the keys will go
rm -rf "$DIR/docker"
mkdir "$DIR/docker"

# copy various types of keys into the host so that our tests can use it
SRC_PATH="/validator-phrases"
DEST_PATH="$DIR/docker/validator-phrases"
CONTAINER=$(docker inspect --format="{{.Id}}" gravity_test_instance)
docker cp "$CONTAINER":"$SRC_PATH" "$DEST_PATH"

SRC_PATH="/orchestrator-phrases"
DEST_PATH="$DIR/docker/orchestrator-phrases"
CONTAINER=$(docker inspect --format="{{.Id}}" gravity_test_instance)
docker cp "$CONTAINER":"$SRC_PATH" "$DEST_PATH"

SRC_PATH="/validator-eth-keys"
DEST_PATH="$DIR/docker/validator-eth-keys"
CONTAINER=$(docker inspect --format="{{.Id}}" gravity_test_instance)
docker cp "$CONTAINER":"$SRC_PATH" "$DEST_PATH"

# run our rust tests
pushd "$DIR/rust/" && cargo run