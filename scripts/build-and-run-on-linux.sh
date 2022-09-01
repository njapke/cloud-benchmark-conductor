#!/usr/bin/env bash

set -euo pipefail

echo "building..."
export CGO_ENABLED=0
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -extldflags '-static'" -o bin-linux-arm64 "$1"

echo "uploading..."
scp bin-linux-arm64 chris@192.168.64.4:~/bin-linux-arm64

shift
echo "running with '$*'"
function sshkill() {
  echo "sending SIGINT"
  ssh chris@192.168.64.4 'sudo kill -s SIGTERM $(pidof ./bin-linux-arm64)'
}
trap sshkill SIGINT
ssh chris@192.168.64.4 "sudo ./bin-linux-arm64 $*" &
while true; do sleep 1; done
