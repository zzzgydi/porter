FROM registry:3.1.0

USER root
RUN apk add --no-cache gettext openssl

# Save original entrypoint before overwriting
RUN if [ -f /entrypoint.sh ]; then cp /entrypoint.sh /entrypoint-original.sh; fi

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# The base image runs as a non-root user; keep its default rather than hardcoding a name.

ENTRYPOINT ["/entrypoint.sh"]
