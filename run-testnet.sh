#!/bin/bash
set -eux

NODES=3

# Build the Docker container
docker build --build-arg NODES=$NODES -t gravity-base .

# Remove existing container instance
set +e
docker rm -f gravity_test_instance
set -e

# Run new test container instance
docker run --name gravity_test_instance \
--cap-add=NET_ADMIN -p 9090:9090 -p 26657:26657 -p 1317:1317 -p 8545:8545 \
-it gravity-base \
/bin/bash /testnet-scripts/run.sh $NODES