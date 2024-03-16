# The daemon image
FROM debian:bookworm-slim
ARG TARGETOS TARGETARCH
COPY ./build/rocketpool-daemon-${TARGETOS}-${TARGETARCH} /usr/bin/rocketpool-daemon
RUN apt update && \
    apt install ca-certificates -y && \
	# Cleanup
	apt clean && \
        rm -rf /var/lib/apt/lists/*

# Container entry point
ENTRYPOINT ["/usr/bin/rocketpool-daemon"]
