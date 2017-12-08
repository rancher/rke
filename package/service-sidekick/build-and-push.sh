#!/bin/bash

ACCT=${ACCT:-rancher}

docker build -t $ACCT/rke-service-sidekick:0.1.0 .
docker push $ACCT/rke-service-sidekick:0.1.0
