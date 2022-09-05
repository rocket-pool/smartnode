###
# Builder
###


# Start from golang image
FROM golang:1.19-alpine AS builder

# Copy source files
ADD . /src

# Compile & install
WORKDIR /src
RUN apk update
RUN apk add --no-cache build-base linux-headers
RUN apk upgrade
ARG CGO_CFLAGS="-O -D__BLST_PORTABLE__"
RUN go install


###
# Process
###


# Start from Alpine image
FROM alpine:latest

# Add C libraries and updates
RUN apk update
RUN apk add --no-cache libgcc libstdc++
RUN apk upgrade

# Copy binary
COPY --from=builder /go/bin/treegen /treegen
