#!/bin/bash
set -eu

# The gaiad binary
BIN=$1

# JSON array of validator information
# [{
#     mnemonic: "crackle snap pop ... etc",
#     allocation: "10000000000stake,10000000000footoken",
#     stake: "5000000000stake",
# }, ... ]
VALIDATORS=$2

# The chain ID
CHAIN_ID=$3

# This is the first 3 fields of the IP addresses which will be used internally by the validators of this blockchain
# Recommended to use something starting with 7, since it is squatted by the DoD and is unroutable on the internet
# For example: "7.7.7"
CHAIN_IP_PREFIX=$4

# A transformation to apply to the genesis file, as a jq string
GENESIS_TRANSFORM=$5

# Whether to skip collecting gentxs so that the genesis does not have them
SKIP_GENTX=$6

# Whether to copy in validator configs from somewhere else
COPY_KEYS=$7



# CREATE VALIDATORS AND DO GENESIS CEREMONY

# Get number of nodes from length of validators array
NODES=$(jq '. | length' <<< "$VALIDATORS")

# first we start a genesis.json with validator0
# validator0 will also collect the gentx's once gnerated
jq -r ".[0].mnemonic" <<< "$VALIDATORS" | $BIN init --home /$CHAIN_ID/validator0 --chain-id=$CHAIN_ID validator0 --recover > /dev/null

# Apply jq transformations to genesis file
jq "$GENESIS_TRANSFORM" /$CHAIN_ID/validator0/config/genesis.json > /$CHAIN_ID/edited-genesis.json
mv /$CHAIN_ID/edited-genesis.json /$CHAIN_ID/genesis.json



# CREATE VALIDATOR HOME FOLDERS ETC

for i in $(seq 0 $(($NODES - 1)));
do
    # TODO: we need to pass in an identifier to identify the validator folder and other things instead of 
    # using the index. The current code WILL BREAK if you supply validators in any other order than 0,1,2 etc
    
    # make the folders for this validator
    mkdir -p /$CHAIN_ID/validator$i/config/

    # Generate an application key for each validator
    # Sets up an arbitrary number of validators on a single machine by manipulating
    # the --home parameter on gaiad
    echo "NODE NUMMMMBBBBEBEBRBERBEBREBRBERER"
    # echo $i
    jq -r ".[$i].mnemonic" <<< "$VALIDATORS"

    jq -r ".[$i].mnemonic" <<< "$VALIDATORS" | $BIN keys add validator$i \
        --home /$CHAIN_ID/validator$i \
        --keyring-backend test \
        --recover > /dev/null
    
    echo "validator$i keys:"
    $BIN keys show validator$i \
        --home /$CHAIN_ID/validator$i \
        --keyring-backend test \
    
    # Give validators their initial token allocations
    # move the genesis in
    mv /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json
    
    # give this validator some money
    ALLOCATION=$(jq -r ".[$i].allocation" <<< "$VALIDATORS")
    $BIN add-genesis-account validator$i $ALLOCATION \
        --home /$CHAIN_ID/validator$i \
        --keyring-backend test

    # move the genesis back out
    mv /$CHAIN_ID/validator$i/config/genesis.json /$CHAIN_ID/genesis.json
done

# echo "BEFORE GENTXS"
# find /$CHAIN_ID/ -print

for i in $(seq 0 $(($NODES - 1)));
do
    # Copy in the genesis.json
    cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json

    # Make a gentx (this command also sets up validator state on disk even if we are not going to use the gentx for anything)
    STAKE_AMOUNT=$(jq -r ".[$i].stake" <<< "$VALIDATORS")
    $BIN gentx validator$i "$STAKE_AMOUNT" \
        --home /$CHAIN_ID/validator$i \
        --keyring-backend test \
        --moniker validator$i \
        --chain-id=$CHAIN_ID \
        --ip $CHAIN_IP_PREFIX.$i
    
    # Copy gentxs to validator0 for possible future collection. 
    # Obviously we don't need to copy validator0's gentx to itself
    if [ $i -gt 0 ]; then
        cp /$CHAIN_ID/validator$i/config/gentx/* /$CHAIN_ID/validator0/config/gentx/
    fi

    # Copy in keys from another chain. This is used to start a consumer chain
    if [ "$COPY_KEYS" != "" ] ; then 
        cp /$COPY_KEYS/validator$i/config/priv_validator_key.json /$CHAIN_ID/validator$i/config/
        cp /$COPY_KEYS/validator$i/config/node_key.json /$CHAIN_ID/validator$i/config/
    fi
done

# echo "AFTER GENTXS"
# find /$CHAIN_ID/ -print




# COLLECT GENTXS IF WE ARE STARTING A NEW CHAIN

if [ "$SKIP_GENTX" = "false" ] ; then 
    # make the final genesis.json
    $BIN collect-gentxs --home /$CHAIN_ID/validator0

    # and copy it to the root 
    cp /$CHAIN_ID/validator0/config/genesis.json /$CHAIN_ID/genesis.json

    # put the now final genesis.json into the correct folders
    for i in $(seq 1 $(($NODES - 1)));
    do
        cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json
    done
fi




# START VALIDATOR NODES

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

    PERSISTENT_PEERS=""

    for j in $(seq 0 $(($NODES - 1)));
    do
        if [ $i -ne $j ]; then
            NODE_ID=$($BIN tendermint show-node-id --home /$CHAIN_ID/validator$j)
            ADDRESS="$NODE_ID@$CHAIN_IP_PREFIX.$j:26656"
            # (jq -r '.body.memo' /$CHAIN_ID/validator$j/config/gentx/*) # Getting the address from the gentx should also work
            PERSISTENT_PEERS="$PERSISTENT_PEERS,$ADDRESS"
        fi
    done

    # Remove leading comma and concat to flag
    PERSISTENT_PEERS="--p2p.persistent_peers ${PERSISTENT_PEERS:1}"

    ARGS="$GAIA_HOME $LISTEN_ADDRESS $RPC_ADDRESS $GRPC_ADDRESS $LOG_LEVEL $P2P_ADDRESS $ENABLE_WEBGRPC $PERSISTENT_PEERS"
    $BIN $ARGS start &> /$CHAIN_ID/validator$i/logs &
done




# poll for chain start
set +e
until interchain-securityd query block --node "tcp://$CHAIN_IP_PREFIX.0:26658" | grep -q -v '{"block_id":{"hash":"","parts":{"total":0,"hash":""}},"block":null}'; do sleep 0.3 ; done
set -e

echo "done!!!!!!!!"

read -p "Press Return to Close..."