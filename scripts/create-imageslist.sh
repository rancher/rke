#!/bin/bash
# This script will create a txt file with RKE extended life images.
set -e -x

IMAGES_LIST_FILE="./build/bin/rke-extended-life-images.txt"

# Ensure the build directory exists
mkdir -p ./build/bin

echo "Creating ${IMAGES_LIST_FILE}"

# The command to generate a unique list of sorted images after removing noiro/aci images. 
./bin/rke config --system-images --all | grep -Eo '^[^ ]+/.+' | grep -v '^noiro/' | sort | uniq > "$IMAGES_LIST_FILE"

# Check if the command was successful
if [ $? -ne 0 ]; then
  echo "Non-zero exit code while running 'rke config --system-images --all'"
  exit 1
fi

echo "Done creating ${IMAGES_LIST_FILE}"

# Display the content of the generated file
cat "$IMAGES_LIST_FILE"
