#!/bin/bash
# If -e is not set then if the build fails, it will use the old container, resulting in a very confusing debugging situation
# Setting -e makes it error out if the build fails
set -eux 

CONTAINER_NAME=$1
INSTANCE_NAME=$2

# Remove existing container instance
set +e
docker rm -f "$INSTANCE_NAME"
set -e

# Build the Docker container
docker build -t "$CONTAINER_NAME" .

# Run new test container instance
docker run --name "$INSTANCE_NAME" --cap-add=NET_ADMIN "$CONTAINER_NAME" /bin/bash /testnet-scripts/beacon.sh