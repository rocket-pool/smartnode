# The daemon image
FROM debian:bookworm-slim
ARG TARGETARCH
COPY ./build/rocketpool-daemon-linux-${TARGETARCH} /usr/local/bin/rocketpool-daemon
RUN apt update && \
    apt install ca-certificates -y && \
	# Cleanup
	apt clean && \
        rm -rf /var/lib/apt/lists/*

# Container entry point
ENTRYPOINT ["/usr/local/bin/rocketpool-daemon"]
