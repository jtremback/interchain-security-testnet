#!/bin/bash
set -eux

BIN=$1
MNEMONICS=$2
CHAIN_ID=$3

# This is the first 3 fields of the IP addresses which will be used internally by the validators of this blockchain
# Recommended to use something starting with 7, since it is squatted by the DoD and is unroutable on the internet
# For example: "7.7.7"
CHAIN_IP_PREFIX=$4

# Default: 26657
RPC_PORT=$5
# Default: 9090
GRPC_PORT=$6

# Get number of nodes from length of mnemonics array
NODES=$(jq '. | length' <<< "$MNEMONICS")

for i in $(seq 0 $(($NODES - 1)));
do
    # add this ip for loopback dialing
    ip addr add $CHAIN_IP_PREFIX.$i/32 dev eth0 || true # allowed to fail

    GAIA_HOME="--home /$CHAIN_ID/validator$i"
    RPC_ADDRESS="--rpc.laddr tcp://$CHAIN_IP_PREFIX.$i:26658"
    GRPC_ADDRESS="--grpc.address $CHAIN_IP_PREFIX.$i:9091"
    LISTEN_ADDRESS="--address tcp://$CHAIN_IP_PREFIX.$i:26655"
    P2P_ADDRESS="--p2p.laddr tcp://$CHAIN_IP_PREFIX.$i:26656"
    LOG_LEVEL="--log_level info"
    ENABLE_WEBGRPC="--grpc-web.enable=false"
    PERSISTENT_PEERS="--p2p.persistent_peers $(paste -sd ',' <<< $(jq -r '.body.memo' /$CHAIN_ID/validator0/config/gentx/*))"

    ARGS="$GAIA_HOME $LISTEN_ADDRESS $RPC_ADDRESS $GRPC_ADDRESS $LOG_LEVEL $P2P_ADDRESS $ENABLE_WEBGRPC $PERSISTENT_PEERS"
    $BIN $ARGS start &> /$CHAIN_ID/validator$i/logs &
done
