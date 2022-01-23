#!/bin/bash
set -eux

BIN=$1
MNEMONICS=$2
CHAIN_ID=$3
CHAIN_IP_PREFIX=$4
GENESIS_TRANSFORM=$5

# Get number of nodes from length of mnemonics array
NODES=$(jq '. | length' <<< "$MNEMONICS")

ALLOCATION="10000000000stake,10000000000footoken"

# first we start a genesis.json with validator 1
# validator 1 will also collect the gentx's once gnerated
# todo add git hash to chain name
jq -r ".[0]" <<< "$MNEMONICS" | $BIN init --home /$CHAIN_ID/validator1 --chain-id=$CHAIN_ID validator1 --recover > /dev/null


## Modify generated genesis.json to our liking by editing fields using jq
## we could keep a hardcoded genesis file around but that would prevent us from
## testing the generated one with the default values provided by the module.

# Apply transformations to genesis file
jq "$GENESIS_TRANSFORM" /$CHAIN_ID/validator1/config/genesis.json > /$CHAIN_ID/edited-genesis.json

mv /$CHAIN_ID/edited-genesis.json /$CHAIN_ID/genesis.json


# Sets up an arbitrary number of validators on a single machine by manipulating
# the --home parameter on gaiad
for i in $(seq 1 $NODES);
do
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


for i in $(seq 1 $NODES);
do
    cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json

    $BIN gentx validator$i 500000000stake --home /$CHAIN_ID/validator$i --keyring-backend test --moniker validator$i --chain-id=$CHAIN_ID --ip $CHAIN_IP_PREFIX.$i
    # obviously we don't need to copy validator1's gentx to itself
    if [ $i -gt 1 ]; then
        cp /$CHAIN_ID/validator$i/config/gentx/* /$CHAIN_ID/validator1/config/gentx/
    fi
done


# make the final genesis.json
$BIN collect-gentxs --home /$CHAIN_ID/validator1

# and copy it to the root 
cp /$CHAIN_ID/validator1/config/genesis.json /$CHAIN_ID/genesis.json

# put the now final genesis.json into the correct folders
for i in $(seq 1 $NODES);
do
    cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json
done

