#!/bin/bash
set -eux


# the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

VAR="bi bim bap"

/bin/bash $DIR/echo-test.sh $VAR