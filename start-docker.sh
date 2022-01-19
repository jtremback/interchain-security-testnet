#!/bin/bash
set -eux

CONTAINER_NAME=$1
INSTANCE_NAME=$2
# Must be in this format "-p 9090:9090 -p 26657:26657 -p 1317:1317 -p 8545:8545"
EXPOSE_PORTS=$3

# Build the Docker container
docker build -t "$CONTAINER_NAME" .

# Remove existing container instance
set +e
docker rm -f "$INSTANCE_NAME"
set -e

# Run new test container instance
docker run --name "$INSTANCE_NAME" --cap-add=NET_ADMIN $EXPOSE_PORTS "$CONTAINER_NAME" /bin/bash /testnet-scripts/beacon.sh