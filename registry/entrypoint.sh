#!/bin/sh
set -e

# Render the config template with environment variables.
if [ -f /etc/distribution/config.yml.template ]; then
    envsubst < /etc/distribution/config.yml.template > /etc/distribution/config.yml
fi

# registry:3 uses "registry serve <config>" (the base image's ENTRYPOINT is
# "registry" and its CMD is "serve /etc/distribution/config.yml").
exec registry serve /etc/distribution/config.yml
