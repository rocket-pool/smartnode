# Start from Alpine image
FROM alpine:latest

# Install rsync
RUN apk add rsync

# Copy the provisioning script
COPY ./ec_migrate.sh /srv

# Container entry point
ENTRYPOINT ["/srv/ec_migrate.sh"]
