# Image for building Smart Node debian packages
FROM debian:bookworm-slim

# Add backports and the unstable repo
RUN cat <<'EOF' > /etc/apt/sources.list
deb http://deb.debian.org/debian bookworm-backports main contrib non-free-firmware
deb-src http://deb.debian.org/debian bookworm-backports main contrib non-free-firmware

deb http://http.us.debian.org/debian unstable main non-free contrib
deb-src http://http.us.debian.org/debian unstable main non-free contrib
EOF

# Install dependencies
RUN apt update && \
    apt install -y -t bookworm devscripts lintian binutils-x86-64-linux-gnu binutils-aarch64-linux-gnu && \
    apt install -y -t bookworm-backports golang-any && \
    apt install -y -t unstable dh-golang && \
	# Cleanup
	apt clean && \
        rm -rf /var/lib/apt/lists/*
