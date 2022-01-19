#!/bin/bash
set -eux

# Build the Docker container
docker build -t is-testnet .

# Remove existing container instance
set +e
docker rm -f is_testnet_instance
set -e

# Run new test container instance
docker run --name is_test_instance \
--cap-add=NET_ADMIN -p 9090:9090 -p 26657:26657 -p 1317:1317 -p 8545:8545 \
-it is-testnet