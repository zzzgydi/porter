FROM registry:3.1.0

# gettext provides envsubst, used by the entrypoint to render the config template.
RUN apk add --no-cache gettext openssl

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
