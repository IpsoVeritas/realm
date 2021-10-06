#!/bin/sh

set -x
set -e

NAMESPACE=${NAMESPACE:-"default"}

key="./realm.pem"

exists=1
kubectl -n ${NAMESPACE} get secret realm-key || exists=0
if [ $exists -gt 0 ]; then
    echo "Key already exists"
    exit 0
fi

./createKey

kubectl -n ${NAMESPACE} create secret generic realm-key --from-file=${key}