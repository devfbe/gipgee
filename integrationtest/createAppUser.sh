#!/bin/sh
set -e

if [ "$1" = "myAlpine-non-root" ]; then
  adduser -D -h /app app
elif [ "$1" = "myUBI-non-root" ]; then
  microdnf install shadow-utils
  adduser -d /app app
  microdnf remove shadow-utils
elif [ "$1" = "myDebian-non-root" ]; then
  useradd -m -d /app app
elif [ "$1" = "myUbuntu-non-root" ]; then
  useradd -m -d /app app
fi
