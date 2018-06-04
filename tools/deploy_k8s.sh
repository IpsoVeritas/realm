#!/bin/bash

# fail if we encounter an error
set -e

export SERVICE="realm-ng"

# make sure we always kill the cloud sql proxy on script exit
function finish {
    # kill the cloud sql proxy
    pkill cloud_sql_proxy || true
}
trap finish EXIT

# Assert that we have all environment variables set
test -n "$VAULT_ADDR" || { echo "ERROR: Environment not setup for vaultenv"; exit 1; }
test -n "$GCE_PROJECT" || { echo "ERROR: Need to specify GCE_PROJECT"; exit 1; }
test -n "$ZONE" || { echo "ERROR: Need to specify ZONE"; exit 1; }
test -n "$CLUSTER" || { echo "ERROR: Need to specify CLUSTER"; exit 1; }
test -n "$NAMESPACE" || { echo "ERROR: Need to specify NAMESPACE"; exit 1; }
test -n "$TAG" || { echo "ERROR: Need to specify TAG"; exit 1; }
test -n "$SQL_NAME" || { echo "ERROR: Need to specify SQL_NAME"; exit 1; }
test -n "$ENV_NAME" || { echo "ERROR: Need to specify ENV_NAME"; exit 1; }

# get GKE credentials
mkdir ~/.kube; vaultenv download secret/env/gitlab-ci/kubeauth-${GCE_PROJECT}.json ~/.kube/gke.json
gcloud auth activate-service-account --key-file ~/.kube/gke.json
gcloud config set container/use_application_default_credentials true
gcloud config set project ${GCE_PROJECT}
export GOOGLE_APPLICATION_CREDENTIALS=~/.kube/gke.json

# get Kubernetes credentials for the GKE cluster
gcloud container clusters get-credentials --zone ${ZONE} ${CLUSTER}

# make sure the namespace exists
kubectl create ns ${NAMESPACE} || true

# make sure we have the correct image pull secret
kubectl -n ${NAMESPACE} get secret gcr-pull || ./tools/gcr_secret.sh

# open the sql proxy towards our CloudSQL instance
export SQL_CONNECT="${GCE_PROJECT}:${REGION}:${SQL_NAME}=tcp:5432"
cloud_sql_proxy -dir ./ -projects ${GCE_PROJECT} -instances=${SQL_CONNECT} & sleep 5

# get SQL credentials from vault
export PGPASSWORD=`vaultenv password secret/env/db/${SQL_NAME}/root`
export SQL_PASSWORD=`vaultenv password secret/env/db/${SQL_NAME}/${SERVICE}`

# make sure database and user exists
psql --user postgres --host 127.0.0.1 -c "CREATE DATABASE ${SERVICE};" || true
psql --user postgres --host 127.0.0.1 -c "CREATE USER ${SERVICE} WITH PASSWORD '${SQL_PASSWORD}';" || true
psql --user postgres --host 127.0.0.1 -c "GRANT ALL PRIVILEGES ON DATABASE ${SERVICE} TO ${SERVICE};"

# make sure the GCE credentials exist as secret in the cluster
vaultenv download secret/env/gitlab-ci/kubeauth-${GCE_PROJECT}.json gce.json
kubectl -n ${NAMESPACE} get secret gce || kubectl -n ${NAMESPACE} create secret generic gce --from-file=gce.json

# get the Key Encryption Key and make sure it exists as a secret
export KEK=`vaultenv password secret/env/${ENV_NAME}/${SERVICE}/kek`
kubectl -n ${NAMESPACE} get secret ${SERVICE}-kek || kubectl -n ${NAMESPACE} create secret generic ${SERVICE}-kek --from-literal=KEK=${KEK}

if [ -f tools/k8s_local.sh ]; then
    ./tools/k8s_local.sh
fi

# setup redis if needed
if [ -f k8s/redis.yml ]; then
    kubectl -n ${NAMESPACE} apply -f k8s/redis.yml
fi

# render and apply
dotenv render -t k8s/pod.yml -o pod.yml
kubectl -n ${NAMESPACE} apply -f pod.yml
dotenv render -t k8s/svc.yml -o svc.yml
kubectl -n ${NAMESPACE} apply -f svc.yml
dotenv render -t k8s/ingress.yml -o ingress.yml
kubectl -n ${NAMESPACE} apply -f ingress.yml

