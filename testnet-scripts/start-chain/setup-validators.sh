#!/bin/bash
set -eux

BIN=$1
NODES=$2
CHAIN_ID=$3

# This is the first 3 fields of the IP addresses which will be used internally by the validators of this blockchain
# Recommended to use something starting with 7, since it is squatted by the DoD and is unroutable on the internet
# For example: "7.7.7"
CHAIN_IP_PREFIX=$4

ALLOCATION="10000000000stake,10000000000footoken"

# first we start a genesis.json with validator 1
# validator 1 will also collect the gentx's once gnerated
STARTING_VALIDATOR=1
STARTING_VALIDATOR_HOME="--home /$CHAIN_ID/validator$STARTING_VALIDATOR"
# todo add git hash to chain name
$BIN init $STARTING_VALIDATOR_HOME --chain-id=$CHAIN_ID validator1


## Modify generated genesis.json to our liking by editing fields using jq
## we could keep a hardcoded genesis file around but that would prevent us from
## testing the generated one with the default values provided by the module.

# a 60 second voting period to allow us to pass governance proposals in the tests
jq '.app_state.gov.voting_params.voting_period = "60s"' /$CHAIN_ID/validator$STARTING_VALIDATOR/config/genesis.json > /$CHAIN_ID/edited-genesis.json

mv /$CHAIN_ID/edited-genesis.json /$CHAIN_ID/genesis.json

ls /$CHAIN_ID/


# Sets up an arbitrary number of validators on a single machine by manipulating
# the --home parameter on gaiad
for i in $(seq 1 $NODES);
do
    GAIA_HOME="--home /$CHAIN_ID/validator$i"
    ARGS="$GAIA_HOME --keyring-backend test"

    # Generate a validator key, orchestrator key, and eth key for each validator
    $BIN keys add $ARGS validator$i 2>> /validator-phrases
    $BIN keys add $ARGS orchestrator$i 2>> /orchestrator-phrases
    # $BIN eth_keys add >> /validator-eth-keys

    VALIDATOR_KEY=$($BIN keys show validator$i -a $ARGS)
    ORCHESTRATOR_KEY=$($BIN keys show orchestrator$i -a $ARGS)
    # move the genesis in
    mkdir -p /$CHAIN_ID/validator$i/config/
    ls /$CHAIN_ID/
    mv /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json
    $BIN add-genesis-account $ARGS $VALIDATOR_KEY $ALLOCATION
    $BIN add-genesis-account $ARGS $ORCHESTRATOR_KEY $ALLOCATION
    # move the genesis back out
    mv /$CHAIN_ID/validator$i/config/genesis.json /$CHAIN_ID/genesis.json
done


for i in $(seq 1 $NODES);
do
    cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json

    $BIN gentx --home /$CHAIN_ID/validator$i --keyring-backend test --moniker validator$i --chain-id=$CHAIN_ID --ip $CHAIN_IP_PREFIX.$i validator$i 500000000stake
    # obviously we don't need to copy validator1's gentx to itself
    if [ $i -gt 1 ]; then
        cp /$CHAIN_ID/validator$i/config/gentx/* /$CHAIN_ID/validator1/config/gentx/
    fi
done


$BIN collect-gentxs $STARTING_VALIDATOR_HOME
GENTXS=$(ls /$CHAIN_ID/validator1/config/gentx | wc -l)
cp /$CHAIN_ID/validator1/config/genesis.json /$CHAIN_ID/genesis.json
echo "Collected $GENTXS gentx"

# put the now final genesis.json into the correct folders
for i in $(seq 1 $NODES);
do
    cp /$CHAIN_ID/genesis.json /$CHAIN_ID/validator$i/config/genesis.json
done
