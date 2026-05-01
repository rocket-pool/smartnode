#!/bin/sh

/usr/share/yum-metrics.sh | sponge /var/lib/node_exporter/textfile_collector/yum.prom || true