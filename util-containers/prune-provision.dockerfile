# Start from Alpine image
FROM alpine:latest

# Copy the provisioning script
COPY ./prune_provision.sh /srv

# Container entry point
ENTRYPOINT ["/srv/prune_provision.sh"]
