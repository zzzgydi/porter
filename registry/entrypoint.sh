#!/bin/sh
set -e

# Generate config from template if template exists
if [ -f /etc/distribution/config.yml.template ]; then
    envsubst < /etc/distribution/config.yml.template > /etc/distribution/config.yml
fi

# Run original registry entrypoint if available, otherwise run registry directly
if [ -f /entrypoint-original.sh ]; then
    exec /entrypoint-original.sh /etc/distribution/config.yml "$@"
else
    exec registry serve /etc/distribution/config.yml "$@"
fi
