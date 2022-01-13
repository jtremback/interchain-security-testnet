#!/bin/bash
set -eux

# Number of validators to start
NODES=$1

# this directy of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

/bin/bash "$DIR/start-validators.sh" "$NODES"

# This keeps the script open to prevent Docker from stopping the container
# immediately if the nodes are killed by a different process
read -p "Press Return to Close..."