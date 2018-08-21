#!/bin/bash

set -e

vaultenv download secret/env/gitlab-ci/mailgun.yml mailgun.yml
kubectl -n ${NAMESPACE} get secret mailgun || kubectl -n ${NAMESPACE} create secret generic mailgun --from-file=mailgun.yml