#!/bin/bash
set -eux

# the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# The gaiad binary
BIN=$1

# Mnemonics with which to start nodes
MNEMONICS=$2

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
GENESIS_TRANSFORM=$7

# How much coin to give each validator on start
# Default: "10000000000stake,10000000000footoken"
ALLOCATION=$8

# Amount for each validator to stake
STAKE_AMOUNT=$9

# Whether to skip collecting gentxs so that the genesis does not have them
SKIP_GENTX=${10}

# generate accounts and do genesis ceremony 
/bin/bash "$DIR/setup-validators.sh" "$BIN" "$MNEMONICS" "$CHAIN_ID" "$CHAIN_IP_PREFIX" "$GENESIS_TRANSFORM" "$ALLOCATION" "$STAKE_AMOUNT" "$SKIP_GENTX"
/bin/bash "$DIR/start-validators.sh" "$BIN" "$MNEMONICS" "$CHAIN_ID" "$CHAIN_IP_PREFIX" "$RPC_PORT" "$GRPC_PORT"

# poll for chain start
set +e
until interchain-securityd query block --node "tcp://$CHAIN_IP_PREFIX.0:26658" | grep -q -v '{"block_id":{"hash":"","parts":{"total":0,"hash":""}},"block":null}'; do sleep 0.3 ; done
set -e

echo "done!!!!!!!!"

read -p "Press Return to Close..."