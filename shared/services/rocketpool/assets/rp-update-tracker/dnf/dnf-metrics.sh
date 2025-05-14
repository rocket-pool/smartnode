#!/bin/sh

# Expose DNF updates and restart info to Prometheus.
#
# Based on yum.sh by Slawomir Gonet <slawek@otwiera.cz>
# and apt.sh by Ben Kochie <superq@gmail.com>
# (see https://github.com/prometheus-community/node-exporter-textfile-collector-scripts/blob/master/yum.sh)

set -u -o pipefail

# shellcheck disable=SC2016
filter_awk_script='
BEGIN { mute=1 }
/Obsoleting Packages/ {
  mute=0
}
mute && /^[[:print:]]+\.[[:print:]]+/ {
  print $3
}
'

check_upgrades() {
  /usr/bin/dnf -q check-update |
    /usr/bin/xargs -n3 |
    awk "${filter_awk_script}" |
    sort |
    uniq -c |
    awk '{print "os_upgrades_pending{origin=\""$2"\"} "$1}'
}

upgrades=$(check_upgrades)

REBOOT=$(needs-restarting -r > /dev/null 2>&1 ; echo "$?")

echo '# HELP os_upgrades_pending DNF package pending updates by origin.'
echo '# TYPE os_upgrades_pending gauge'
if [[ -n "${upgrades}" ]] ; then
  echo "${upgrades}"
else
  echo 'os_upgrades_pending{origin=""} 0'
fi

echo '# HELP os_reboot_required Node reboot is required for software updates.'
echo '# TYPE os_reboot_required gauge'
echo "os_reboot_required ${REBOOT}"
