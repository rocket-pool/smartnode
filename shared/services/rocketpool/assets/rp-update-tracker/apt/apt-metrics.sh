#!/bin/sh

if [ -f /usr/lib/update-notifier/apt-check ]; then
    # For Ubuntu systems
    APT_CHECK=$(/usr/lib/update-notifier/apt-check 2>&1 || echo "0;0")
    UPDATES=$(echo "$APT_CHECK" | cut -d ';' -f 1)
    SECURITY=$(echo "$APT_CHECK" | cut -d ';' -f 2)
else
    # For Debian systems
    UPDATES=$(LANG=C apt-get dist-upgrade -s | grep -P '^\d+ upgraded'| cut -d" " -f1)
    SECURITY=0
fi

REBOOT=$([ -f /var/run/reboot-required ] && echo 1 || echo 0)

echo "# HELP os_upgrades_pending Apt package pending updates by origin."
echo "# TYPE os_upgrades_pending gauge"
echo "os_upgrades_pending ${UPDATES}"

echo "# HELP os_security_upgrades_pending Apt package pending security updates by origin."
echo "# TYPE os_security_upgrades_pending gauge"
echo "os_security_upgrades_pending ${SECURITY}"

echo "# HELP os_reboot_required Node reboot is required for software updates."
echo "# TYPE os_reboot_required gauge"
echo "os_reboot_required ${REBOOT}"
