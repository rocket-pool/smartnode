#!/bin/sh

/usr/share/yum-metrics.sh | sponge /var/lib/node_exporter/textfile_collector/yum.prom || true
/usr/share/rp-version-check.sh | sponge /var/lib/node_exporter/textfile_collector/rp.prom || true