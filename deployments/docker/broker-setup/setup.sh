#!/usr/bin/env bash

#
# Copyright 2021 - now, the original author or authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e -x

PULSAR_VERSION=${PULSAR_VERSION:-2.8.0}
PULSAR_CONNECTORS=${PULSAR_CONNECTORS:-""}
PULSAR_DATA_DIRECTORY=${PULSAR_DATA_DIRECTORY:-$(pwd)}
PULSAR_CONNECTORS_DIRECTORY="$PULSAR_DATA_DIRECTORY/connectors"
PULSAR_CONNECTORS_BASE_URL=${PULSAR_CONNECTORS_BASE_URL:-"https://archive.apache.org/dist/pulsar"}

mkdir -p "$PULSAR_CONNECTORS_DIRECTORY"

function getFirstPart() {
  local str=$1
  first=${str%;*}
  echo "$first"
  return
}

function getSecondPart() {
  str=$1
  if [[ "$str" == *";"* ]]; then
    echo "$1" | cut -d ";" -f2
  fi
  return
}

function getConnectorName() {
  url=$1
  # shellcheck disable=SC2001
  echo "$url" | sed 's:.*/::'
  return
}

function getConnectorUrl() {
  part1=$(getFirstPart "$1")
  if [[ "$part1" =~ ^(https?)://(.*) ]]; then
    echo "$part1"
  else
    echo "${PULSAR_CONNECTORS_BASE_URL}/pulsar-${PULSAR_VERSION}/connectors/pulsar-io-${part1}-${PULSAR_VERSION}.nar"
  fi
  return
}

function generateCurlHeaders() {
  headers=""
  part2=$(getSecondPart "$1")
  IFS=',' read -r -a keyValues <<<"$part2"
  for hder in "${keyValues[@]}"; do
    headers+=" -H \"$hder\""
  done
  echo "$headers"
  return
}

IFS=' ' read -r -a connectors <<<"$PULSAR_CONNECTORS"

for connector in "${connectors[@]}"; do
  url=$(getConnectorUrl "$connector")
  name=$(getConnectorName "$url")

  headers=$(generateCurlHeaders "$connector")
  if [[ -f "$PULSAR_CONNECTORS_DIRECTORY/$name" ]]; then
    printf "The connector %s already exists, Skipping...\n" "$name"
    continue
  fi
  printf "Downloading the connector: %s from %s\n" "$name" "$url"
  if [[ ! "$headers" == "" ]]; then
    curl "$headers" "$url" -o "$name"
  else
    curl "$url" -o "$name"
  fi
  # shellcheck disable=SC2181
  if [[ $? -ne 0 ]]; then
    printf "Unable to download the connector: %s\n" "$url"
    exit 1
  fi
  printf "Download successful; moving %s to %s\n" "$name" "$PULSAR_CONNECTORS_DIRECTORY"
  mv "$name" "$PULSAR_CONNECTORS_DIRECTORY"
done

printf "Setup is successfully. ✨✨\n"
