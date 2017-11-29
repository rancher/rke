#!/bin/bash

ACCT=${ACCT:-rancher}

docker build -t $ACCT/rke-nginx-proxy:0.1.0 .
docker push $ACCT/rke-nginx-proxy:0.1.0
