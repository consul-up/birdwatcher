#!/usr/bin/env bash

# This script gets wikipedia's JSON for each bird in birds.txt/canaries.txt
# and outputs it to birds.json and canaries.json. You then need to add `[` and `]` to that
# file and remove the trailing comma to get a proper JSON array.
while read p; do
  curl -s -L "$p" | jq -c . >> birds.json && echo ',' >> birds.json
done <birds.txt

while read p; do
  curl -s -L "$p" | jq -c . >> canaries.json && echo ',' >> canaries.json
done <canaries.txt
