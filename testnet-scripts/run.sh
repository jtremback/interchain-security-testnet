#!/bin/bash
set -eux

# the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# /bin/bash "$DIR/start-chain/start-chain.sh" interchain-securityd 3 provider 7.7.7 26657 9090 '.app_state.gov.voting_params.voting_period = "60s"'

# /bin/bash "$DIR/start-chain/start-chain.sh" interchain-securityd 3 consumer 7.7.8 26757 9190 '.app_state.gov.voting_params.voting_period = "60s"'

# This keeps the script open to prevent Docker from stopping the container
# immediately if the nodes are killed by a different process
read -p "Press Return to Close..."