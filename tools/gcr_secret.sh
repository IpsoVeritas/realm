#!/bin/bash

# exit on failure
set -e

test -n "$NAMESPACE" || { echo "ERROR: Need to specify NAMESPACE"; exit 1; }

# download secret
vaultenv download secret/env/gitlab-ci/gcr-pull.json gcr-pull.json

# create secret in kubernetes
kubectl -n ${NAMESPACE} create secret docker-registry gcr-pull \
  --docker-server=https://gcr.io \
  --docker-username=_json_key \
  --docker-email=user@example.com \
  --docker-password="$(cat gcr-pull.json)"

kubectl -n ${NAMESPACE} patch serviceaccount default -p "{\"imagePullSecrets\": [{\"name\": \"gcr-pull\"}]}"