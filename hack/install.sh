#!/bin/bash

set -e

new_deployment='./artifacts/deployment.yaml'
template_deployment='./artifacts/deployment-template.yaml'

cp -f $template_deployment $new_deployment
./hack/generate.sh ${new_deployment}

kubectl --context=$1 apply -f ${new_deployment}
kubectl --context=$1 get pods -n central | grep sqsmv
