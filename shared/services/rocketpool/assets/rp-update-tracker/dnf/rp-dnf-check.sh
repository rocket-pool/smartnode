#!/bin/sh

/usr/share/dnf-metrics.sh | sponge /var/lib/node_exporter/textfile_collector/dnf.prom || true
/usr/share/rp-version-check.sh | sponge /var/lib/node_exporter/textfile_collector/rp.prom || true