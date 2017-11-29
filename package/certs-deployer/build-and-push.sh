#!/bin/bash

ACCT=${ACCT:-rancher}

docker build -t $ACCT/rke-cert-deployer:0.1.0 .
docker push $ACCT/rke-cert-deployer:0.1.0
