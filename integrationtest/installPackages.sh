#!/bin/sh
set -e
if [ "$1" = "myAlpine" ]; then
    apk --no-cache upgrade
    apk add curl wget
elif [ "$1" = "myUBI" ]; then
    microdnf upgrade
    microdnf install curl wget
elif [ "$1" = "myUbuntu" ] || [ "$1" = "myDebian" ]; then
    apt-get update && apt-get -y dist-upgrade && apt-get -y install curl wget
fi
