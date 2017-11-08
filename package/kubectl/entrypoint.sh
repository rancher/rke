#!/bin/bash -x

# Set template configration
for i in $(env | grep -o RKE_.*=); do
  key=$(echo "$i" | cut -f1 -d"=")
  value=$(echo "${!key}")
  for f in /network/*.yaml /addons/*.yaml; do
    sed -i "s|${key}|${value}|g" ${f}
  done
done


for i in $(env | grep -o KUBECFG_.*=); do
  name="$(echo "$i" | cut -f1 -d"=" | tr '[:upper:]' '[:lower:]' | tr '_' '-').yaml"
  env=$(echo "$i" | cut -f1 -d"=")
  value=$(echo "${!env}")
  if [ ! -f $SSL_CRTS_DIR/$name ]; then
    echo "$value" > /root/.kube/config
  fi
done

kubectl ${@}
