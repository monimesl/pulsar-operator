#!/usr/bin/env bash

set -e -x

PULSAR_VERSION=${PULSAR_VERSION:-2.8.0}
PULSAR_CONNECTORS=${PULSAR_CONNECTORS:-"aerospike"}
PULSAR_SETUP_DIRECTORY=${PULSAR_SETUP_DIRECTORY:-$(pwd)}
PULSAR_CONNECTORS_DIRECTORY="$PULSAR_SETUP_DIRECTORY/connectors"
PULSAR_CONNECTORS_BASE_URL=${PULSAR_CONNECTORS_BASE_URL:-"https://archive.apache.org/dist/pulsar"}

mkdir -p "$PULSAR_SETUP_DIRECTORY"

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

mkdir -p "$PULSAR_CONNECTORS_DIRECTORY"

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
