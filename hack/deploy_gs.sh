#!/usr/bin/env bash

set -eu
set -o pipefail

action=$1
fileName=$2
gsCount=${3:-1}

for (( i = 0; i < ${gsCount}; ++i )); do
    export GS_NAME=${i}
    cat $fileName | envsubst | kubectl ${action} -f -
done


