#!/bin/bash
# set -eux

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

for i in $(seq 1 $NODES);
do
    # add this ip for loopback dialing
    ip addr add $CHAIN_IP_PREFIX.$i/32 dev eth0 || true # allowed to fail

    GAIA_HOME="--home /$CHAIN_ID/validator$i"
    # this implicitly caps us at ~6000 nodes for this sim
    # note that we start on 26656 the idea here is that the first
    # node (node 1) is at the expected contact address from the gentx
    # faciliating automated peer exchange
    if [[ "$i" -eq 1 ]]; then
        # node one gets localhost so we can easily shunt these ports
        # to the docker host
        RPC_ADDRESS="--rpc.laddr tcp://0.0.0.0:$RPC_PORT"
        GRPC_ADDRESS="--grpc.address 0.0.0.0:$GRPC_PORT"
    else
        # move these to another port and address, not becuase they will
        # be used there, but instead to prevent them from causing problems
        # you also can't duplicate the port selection against localhost
        # for reasons that are not clear to me right now.
        RPC_ADDRESS="--rpc.laddr tcp://$CHAIN_IP_PREFIX.$i:26658"
        GRPC_ADDRESS="--grpc.address $CHAIN_IP_PREFIX.$i:9091"
    fi
    LISTEN_ADDRESS="--address tcp://$CHAIN_IP_PREFIX.$i:26655"
    P2P_ADDRESS="--p2p.laddr tcp://$CHAIN_IP_PREFIX.$i:26656"
    LOG_LEVEL="--log_level info"
    ENABLE_WEBGRPC="--grpc-web.enable=false"

    ARGS="$GAIA_HOME $LISTEN_ADDRESS $RPC_ADDRESS $GRPC_ADDRESS $LOG_LEVEL $P2P_ADDRESS $ENABLE_WEBGRPC"
    $BIN $ARGS start &> /$CHAIN_ID/validator$i/logs &
done
