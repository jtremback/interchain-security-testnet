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





# CREATE VALDIATORS AND DO GENESIS CEREMONY

# Get number of nodes from length of mnemonics array
NODES=$(jq '. | length' <<< "$MNEMONICS")


# first we start a genesis.json with validator0
# validator0 will also collect the gentx's once gnerated
# todo add git hash to chain name
jq -r ".[0]" <<< "$MNEMONICS" | $BIN init --home /$CHAIN_ID/validator0 --chain-id=$CHAIN_ID validator0 --recover > /dev/null


## Modify generated genesis.json to our liking by editing fields using jq
## we could keep a hardcoded genesis file around but that would prevent us from
## testing the generated one with the default values provided by the module.

# Apply transformations to genesis file
jq "$GENESIS_TRANSFORM" /$CHAIN_ID/validator0/config/genesis.json > /$CHAIN_ID/edited-genesis.json

mv /$CHAIN_ID/edited-genesis.json /$CHAIN_ID/genesis.json


# Sets up an arbitrary number of validators on a single machine by manipulating
# the --home parameter on gaiad
for i in $(seq 0 $(($NODES - 1)));
do
    # TODO: we need to pass in an identifier to identify the validator folder and other things instead of 
    # using the index
    
    # make the folders for this validator
    mkdir -p /$CHAIN_ID/validator$i/config/
    
    ARGS="--home /$CHAIN_ID/validator$i --keyring-backend test"

    # Generate a validator key, orchestrator key, and eth key for each validator
    jq -r ".[$((i-1))]" <<< "$MNEMONICS" | $BIN keys add $ARGS validator$i --recover > /$CHAIN_ID/validator$i/mnemonic

    echo "validator$i keys:"
    $BIN keys show validator$i $ARGS

    # move the genesis in
    mv /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json
    $BIN add-genesis-account $ARGS validator$i $ALLOCATION

    # move the genesis back out
    mv /$CHAIN_ID/validator$i/config/genesis.json /$CHAIN_ID/genesis.json
done


for i in $(seq 0 $(($NODES - 1)));
do
    cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json

    $BIN gentx validator$i "$STAKE_AMOUNT" --home /$CHAIN_ID/validator$i --keyring-backend test --moniker validator$i --chain-id=$CHAIN_ID --ip $CHAIN_IP_PREFIX.$i

    # obviously we don't need to copy validator0's gentx to itself
    if [ $i -gt 0 ]; then
        cp /$CHAIN_ID/validator$i/config/gentx/* /$CHAIN_ID/validator0/config/gentx/
    fi
done

if [ "$SKIP_GENTX" = "false" ] ; then 
    # make the final genesis.json
    $BIN collect-gentxs --home /$CHAIN_ID/validator0
fi

# and copy it to the root 
cp /$CHAIN_ID/validator0/config/genesis.json /$CHAIN_ID/genesis.json

# put the now final genesis.json into the correct folders
for i in $(seq 1 $(($NODES - 1)));
do
    cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json
done




# START VALIDATOR NODES

# Set up seed node
$BIN init --home /$CHAIN_ID/seed --chain-id=$CHAIN_ID seed
cp /$CHAIN_ID/genesis.json /$CHAIN_ID/seed/config/genesis.json
SEED_ID=$($BIN tendermint show-node-id --home /$CHAIN_ID/seed)
ip addr add $CHAIN_IP_PREFIX.254/32 dev eth0 || true # allowed to fail
$BIN start \
    --home /$CHAIN_ID/seed \
    --p2p.laddr tcp://$CHAIN_IP_PREFIX.254:26656 \
    --rpc.laddr tcp://$CHAIN_IP_PREFIX.254:26658 \
    --grpc.address $CHAIN_IP_PREFIX.254:9091 \
    --address tcp://$CHAIN_IP_PREFIX.254:26655 \
    --p2p.laddr tcp://$CHAIN_IP_PREFIX.254:26656 \
    --grpc-web.enable=false \
    &> /$CHAIN_ID/seed/logs &

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

    ARGS="$GAIA_HOME $LISTEN_ADDRESS $RPC_ADDRESS $GRPC_ADDRESS $LOG_LEVEL $P2P_ADDRESS $ENABLE_WEBGRPC --p2p.seeds $SEED_ID@$CHAIN_IP_PREFIX.254:26656"
    $BIN $ARGS start &> /$CHAIN_ID/validator$i/logs &
done




# poll for chain start
set +e
until interchain-securityd query block --node "tcp://$CHAIN_IP_PREFIX.0:26658" | grep -q -v '{"block_id":{"hash":"","parts":{"total":0,"hash":""}},"block":null}'; do sleep 0.3 ; done
set -e

echo "done!!!!!!!!"

read -p "Press Return to Close..."