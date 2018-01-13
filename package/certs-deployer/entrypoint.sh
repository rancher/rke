#!/bin/bash -x

SSL_CRTS_DIR=${CRTS_DEPLOY_PATH:-/etc/kubernetes/ssl}
mkdir -p $SSL_CRTS_DIR

for i in $(env | grep -o KUBE_.*=); do
  name="$(echo "$i" | cut -f1 -d"=" | tr '[:upper:]' '[:lower:]' | tr '_' '-').pem"
  env=$(echo "$i" | cut -f1 -d"=")
  value=$(echo "${!env}")
  if [ ! -f $SSL_CRTS_DIR/$name ] || [ "$FORCE_DEPLOY" == "true" ]; then
    echo "$value" > $SSL_CRTS_DIR/$name
  fi
done

for i in $(env | grep -o KUBECFG_.*=); do
  name="$(echo "$i" | cut -f1 -d"=" | tr '[:upper:]' '[:lower:]' | tr '_' '-').yaml"
  env=$(echo "$i" | cut -f1 -d"=")
  value=$(echo "${!env}")
  if [ ! -f $SSL_CRTS_DIR/$name ]; then
    echo "$value" > $SSL_CRTS_DIR/$name
  fi
done
