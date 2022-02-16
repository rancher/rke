#!/bin/bash
# This script will create a txt file with k8s versions which will be used as (pre) release decription by Drone
set -e -x

RELEASEFILE="./build/bin/rke-k8sversions.txt"

mkdir -p ./build/bin

echo "Creating ${RELEASEFILE}"

DEFAULT_VERSION=$(./bin/rke --quiet config --list-version)
if [ $? -ne 0 ]; then
  echo "Non zero exit code while running 'rke config -l'"
  exit 1
fi

DEFAULT_VERSION_FOUND="false"
echo "# RKE Kubernetes versions" > $RELEASEFILE
for VERSION in $(./bin/rke --quiet config --all --list-version | sort -V); do
  if [ "$VERSION" == "$DEFAULT_VERSION" ]; then
    echo "- \`${VERSION}\` (default)" >> $RELEASEFILE
    DEFAULT_VERSION_FOUND="true"
  else
    echo "- \`${VERSION}\`" >> $RELEASEFILE
  fi
done

if [ "$DEFAULT_VERSION_FOUND" == "false" ]; then
  echo -e "\nNo default version found!" >> $RELEASEFILE
fi

echo "Done creating ${RELEASEFILE}"

cat $RELEASEFILE
