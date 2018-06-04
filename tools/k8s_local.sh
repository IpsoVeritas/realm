#!/bin/bash

set -e

#vaultenv download secret/env/gitlab-ci/kubeauth-${GCE_PROJECT}.json gce.json
kubectl -n ${NAMESPACE} get secret mailgun || kubectl -n ${NAMESPACE} create secret generic mailgun --from-file=cmd/realm/dev.yml