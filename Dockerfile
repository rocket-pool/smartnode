FROM rocketpool/smartnode-base

ARG TARGETARCH
COPY ./Releases/treegen-linux-${TARGETARCH} /treegen
