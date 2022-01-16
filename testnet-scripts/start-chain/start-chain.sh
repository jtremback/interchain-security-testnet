#!/bin/bash
set -eux

# the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# The gaiad binary
BIN=$1

# Number of nodes
NODES=$2

# The chain ID
CHAIN_ID=$3

# This is the first 3 fields of the IP addresses which will be used internally by the validators of this blockchain
# Recommended to use something starting with 7, since it is squatted by the DoD and is unroutable on the internet
# For example: "7.7.7"
CHAIN_IP_PREFIX=$4

# Default: 26657
RPC_PORT=$5

# Default: 9090
GRPC_PORT=$6

# A transformation to apply to the genesis file, as a jq string
GENESIS_TRANSFORMATION=$7

/bin/bash "$DIR/setup-validators.sh" "$BIN" "$NODES" "$CHAIN_ID" "$CHAIN_IP_PREFIX" "$GENESIS_TRANSFORMATION"
/bin/bash "$DIR/start-validators.sh" "$BIN" "$NODES" "$CHAIN_ID" "$CHAIN_IP_PREFIX" "$RPC_PORT" "$GRPC_PORT"