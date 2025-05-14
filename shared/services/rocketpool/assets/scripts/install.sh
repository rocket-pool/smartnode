#!/bin/sh

##
# Rocket Pool service installation script
# Prints progress messages to stdout
# All command output is redirected to stderr
##

COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[33m'
COLOR_RESET='\033[0m'

# Print a failure message to stderr and exit
fail() {
    MESSAGE=$1
    >&2 echo -e "\n${COLOR_RED}**ERROR**\n$MESSAGE${COLOR_RESET}"
    exit 1
}


# Get CPU architecture
UNAME_VAL=$(uname -m)
ARCH=""
case $UNAME_VAL in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       fail "CPU architecture not supported: $UNAME_VAL" ;;
esac


# Get the platform type
PLATFORM=$(uname -s)
if [ "$PLATFORM" = "Linux" ]; then
    if command -v lsb_release &>/dev/null ; then
        PLATFORM=$(lsb_release -si)
    elif [ -f "/etc/centos-release" ]; then
        PLATFORM="CentOS"
    elif [ -f "/etc/fedora-release" ]; then
        PLATFORM="Fedora"
    fi
fi

##
# Config
##


# The total number of steps in the installation process
TOTAL_STEPS="8"
# The Rocket Pool user data path
RP_PATH="$HOME/.rocketpool"
# The default network to run Rocket Pool on
NETWORK="mainnet"

##
# Utils
##


# Print progress
progress() {
    STEP_NUMBER=$1
    MESSAGE=$2
    echo "Step $STEP_NUMBER of $TOTAL_STEPS: $MESSAGE"
}


# Docker installation steps
add_user_docker() {
    $SUDO_CMD usermod -aG docker $USER || fail "Could not add user to docker group."
}


# Detect installed privilege escalation programs
get_escalation_cmd() {
    if type sudo > /dev/null 2>&1; then
        SUDO_CMD="sudo"
    elif type doas > /dev/null 2>&1; then
        echo "NOTE: sudo not found, using doas instead"
        SUDO_CMD="doas"
    else
        fail "Please make sure a privilege escalation command such as \"sudo\" is installed and available before installing Rocket Pool."
    fi
}

# Install
install() {


##
# Initialization
##


# Parse arguments
while getopts "dp:u:n:" FLAG; do
    case "$FLAG" in
        d) NO_DEPS=true ;;
        p) RP_PATH="$OPTARG" ;;
        u) DATA_PATH="$OPTARG" ;;
        n) NETWORK="$OPTARG" ;;
        *) fail "Incorrect usage." ;;
    esac
done

if [ -z "$DATA_PATH" ]; then
    DATA_PATH="$RP_PATH/data"
fi

# Get temporary data paths
PACKAGE_FILES_PATH="$(dirname $0)/install"

##
# Installation
##


# OS dependencies
if [ -z "$NO_DEPS" ]; then

>&2 get_escalation_cmd

case "$PLATFORM" in

    # Ubuntu / Debian / Raspbian
    Ubuntu|Debian|Raspbian)

        # Get platform name
        PLATFORM_NAME=$(echo "$PLATFORM" | tr '[:upper:]' '[:lower:]')

        # Install OS dependencies
        progress 1 "Installing OS dependencies..."
        { $SUDO_CMD apt-get -y update || fail "Could not update OS package definitions."; } >&2
        { $SUDO_CMD apt-get -y install apt-transport-https ca-certificates curl gnupg gnupg-agent lsb-release software-properties-common chrony || fail "Could not install OS packages."; } >&2

        # Check for existing Docker installation
        progress 2 "Checking if Docker is installed..."
        dpkg-query -W -f='${Status}' docker-ce 2>&1 | grep -q -P '^install ok installed$' > /dev/null
        if [ $? != "0" ]; then
            echo "Installing Docker..."
            if [ ! -f /etc/apt/sources.list.d/docker.list ]; then
                # Install the Docker repo
                { $SUDO_CMD mkdir -p /etc/apt/keyrings || fail "Could not create APT keyrings directory."; } >&2
                { curl -fsSL "https://download.docker.com/linux/$PLATFORM_NAME/gpg" | $SUDO_CMD gpg --dearmor -o /etc/apt/keyrings/docker.gpg || fail "Could not add docker repository key."; } >&2
                { echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$PLATFORM_NAME $(lsb_release -cs) stable" | $SUDO_CMD tee /etc/apt/sources.list.d/docker.list > /dev/null || fail "Could not add docker repository."; } >&2
            fi
            { $SUDO_CMD apt-get -y update || fail "Could not update OS package definitions."; } >&2
            { $SUDO_CMD apt-get -y install docker-ce docker-ce-cli docker-compose-plugin containerd.io || fail "Could not install Docker packages."; } >&2
        fi

        # Check for existing docker-compose-plugin installation
        progress 2 "Checking if docker-compose-plugin is installed..."
        dpkg-query -W -f='${Status}' docker-compose-plugin 2>&1 | grep -q -P '^install ok installed$' > /dev/null
        if [ $? != "0" ]; then
            echo "Installing docker-compose-plugin..."
            if [ ! -f /etc/apt/sources.list.d/docker.list ]; then
                # Install the Docker repo, removing the legacy one if it exists
                { $SUDO_CMD add-apt-repository --remove "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/$PLATFORM_NAME $(lsb_release -cs) stable"; } 2>/dev/null
                { $SUDO_CMD mkdir -p /etc/apt/keyrings || fail "Could not create APT keyrings directory."; } >&2
                { curl -fsSL "https://download.docker.com/linux/$PLATFORM_NAME/gpg" | $SUDO_CMD gpg --dearmor -o /etc/apt/keyrings/docker.gpg || fail "Could not add docker repository key."; } >&2
                { echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$PLATFORM_NAME $(lsb_release -cs) stable" | $SUDO_CMD tee /etc/apt/sources.list.d/docker.list > /dev/null || fail "Could not add Docker repository."; } >&2
            fi
            { $SUDO_CMD apt-get -y update || fail "Could not update OS package definitions."; } >&2
            { $SUDO_CMD apt-get -y install docker-compose-plugin || fail "Could not install docker-compose-plugin."; } >&2
            { $SUDO_CMD systemctl restart docker || fail "Could not restart docker daemon."; } >&2
        else
            echo "Already installed."
        fi

        # Add user to docker group
        progress 3 "Adding user to docker group..."
        >&2 add_user_docker

    ;;

    # Centos
    CentOS)

        # Install OS dependencies
        progress 1 "Installing OS dependencies..."
        { $SUDO_CMD yum install -y yum-utils chrony || fail "Could not install OS packages."; } >&2
        { $SUDO_CMD systemctl start chronyd || fail "Could not start chrony daemon."; } >&2

        # Install docker
        progress 2 "Installing docker..."
        { $SUDO_CMD yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo || fail "Could not add docker repository."; } >&2
        { $SUDO_CMD yum install -y docker-ce docker-ce-cli docker-compose-plugin containerd.io || fail "Could not install docker packages."; } >&2
        { $SUDO_CMD systemctl start docker || fail "Could not start docker daemon."; } >&2
        { $SUDO_CMD systemctl enable docker || fail "Could not set docker daemon to auto-start on boot."; } >&2

        # Check for existing docker-compose-plugin installation
        progress 2 "Checking if docker-compose-plugin is installed..."
        yum -q list installed docker-compose-plugin 2>/dev/null 1>/dev/null
        if [ $? != "0" ]; then
            echo "Installing docker-compose-plugin..."
            { $SUDO_CMD yum install -y docker-compose-plugin || fail "Could not install docker-compose-plugin."; } >&2
            { $SUDO_CMD systemctl restart docker || fail "Could not restart docker daemon."; } >&2
        else
            echo "Already installed."
        fi

        # Add user to docker group
        progress 3 "Adding user to docker group..."
        >&2 add_user_docker

    ;;

    # Fedora
    Fedora)

        # Install OS dependencies
        progress 1 "Installing OS dependencies..."
        { $SUDO_CMD dnf -y install dnf-plugins-core chrony || fail "Could not install OS packages."; } >&2
        { $SUDO_CMD systemctl start chronyd || fail "Could not start chrony daemon."; } >&2

        # Install docker
        progress 2 "Installing docker..."
        { $SUDO_CMD dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo || fail "Could not add docker repository."; } >&2
        { $SUDO_CMD dnf -y install docker-ce docker-ce-cli docker-compose-plugin containerd.io || fail "Could not install docker packages."; } >&2
        { $SUDO_CMD systemctl start docker || fail "Could not start docker daemon."; } >&2
        { $SUDO_CMD systemctl enable docker || fail "Could not set docker daemon to auto-start on boot."; } >&2

        # Check for existing docker-compose-plugin installation
        progress 2 "Checking if docker-compose-plugin is installed..."
        dnf -q list installed docker-compose-plugin 2>/dev/null 1>/dev/null
        if [ $? != "0" ]; then
            echo "Installing docker-compose-plugin..."
            { $SUDO_CMD dnf install -y docker-compose-plugin || fail "Could not install docker-compose-plugin."; } >&2
            { $SUDO_CMD systemctl restart docker || fail "Could not restart docker daemon."; } >&2
        else
            echo "Already installed."
        fi

        # Add user to docker group
        progress 3 "Adding user to docker group..."
        >&2 add_user_docker

    ;;

    # Unsupported OS
    *)
        RED='\033[0;31m'
        echo ""
        echo -e "${RED}**ERROR**"
        echo "Automatic dependency installation for the $PLATFORM operating system is not supported."
        echo "Please install docker and docker-compose-plugin manually, then try again with the '-d' flag to skip OS dependency installation."
        echo "Be sure to add yourself to the docker group with '$SUDO_CMD usermod -aG docker $USER' after installing docker."
        echo "Log out and back in, or restart your system after you run this command."
        echo -e "${RESET}"
        exit 1
    ;;

esac
else
    echo "Skipping steps 1 - 2 (OS dependencies & docker)"
    case "$PLATFORM" in
        # Ubuntu / Debian / Raspbian
        Ubuntu|Debian|Raspbian)

            # Get platform name
            PLATFORM_NAME=$(echo "$PLATFORM" | tr '[:upper:]' '[:lower:]')

            # Check for existing docker-compose-plugin installation
            progress 3 "Checking if docker-compose-plugin is installed..."
            dpkg-query -W -f='${Status}' docker-compose-plugin 2>&1 | grep -q -P '^install ok installed$' > /dev/null
            if [ $? != "0" ]; then
                >&2 get_escalation_cmd
                echo "Installing docker-compose-plugin..."
                if [ ! -f /etc/apt/sources.list.d/docker.list ]; then
                    # Install the Docker repo, removing the legacy one if it exists
                    { $SUDO_CMD add-apt-repository --remove "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/$PLATFORM_NAME $(lsb_release -cs) stable"; } 2>/dev/null
                    { $SUDO_CMD mkdir -p /etc/apt/keyrings || fail "Could not create APT keyrings directory."; } >&2
                    { curl -fsSL "https://download.docker.com/linux/$PLATFORM_NAME/gpg" | $SUDO_CMD gpg --dearmor -o /etc/apt/keyrings/docker.gpg || fail "Could not add docker repository key."; } >&2
                    { echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$PLATFORM_NAME $(lsb_release -cs) stable" | $SUDO_CMD tee /etc/apt/sources.list.d/docker.list > /dev/null || fail "Could not add Docker repository."; } >&2
                fi
                { $SUDO_CMD apt-get -y update || fail "Could not update OS package definitions."; } >&2
                { $SUDO_CMD apt-get -y install docker-compose-plugin || fail "Could not install docker-compose-plugin."; } >&2
                { $SUDO_CMD systemctl restart docker || fail "Could not restart docker daemon."; } >&2
            else
                echo "Already installed."
            fi

        ;;

        # Centos
        CentOS)

            # Check for existing docker-compose-plugin installation
            progress 3 "Checking if docker-compose-plugin is installed..."
            yum -q list installed docker-compose-plugin 2>/dev/null 1>/dev/null
            if [ $? != "0" ]; then
                >&2 get_escalation_cmd
                echo "Installing docker-compose-plugin..."
                { $SUDO_CMD yum install -y docker-compose-plugin || fail "Could not install docker-compose-plugin."; } >&2
                { $SUDO_CMD systemctl restart docker || fail "Could not restart docker daemon."; } >&2
            else
                echo "Already installed."
            fi

        ;;

        # Fedora
        Fedora)

            # Check for existing docker-compose-plugin installation
            progress 3 "Checking if docker-compose-plugin is installed..."
            dnf -q list installed docker-compose-plugin 2>/dev/null 1>/dev/null
            if [ $? != "0" ]; then
                >&2 get_escalation_cmd
                echo "Installing docker-compose-plugin..."
                { $SUDO_CMD dnf install -y docker-compose-plugin || fail "Could not install docker-compose-plugin."; } >&2
                { $SUDO_CMD systemctl restart docker || fail "Could not restart docker daemon."; } >&2
            else
                echo "Already installed."
            fi

        ;;

        # Everything else
        *)
            # Check for existing docker-compose-plugin installation
            progress 3 "Checking if docker-compose-plugin is installed..."
            if docker compose 2>/dev/null 1>/dev/null ; then
                echo "Already installed."
            else
                RED='\033[0;31m'
                echo ""
                echo -e "${RED}**ERROR**"
                echo "The docker-compose-plugin package is not installed. Starting with v1.7.0, the Smartnode requires this package because the legacy docker-compose script is no longer supported."
                echo "Since automatic dependency installation for the $PLATFORM operating system is not supported, you will need to install it manually."
                echo "Please install docker-compose-plugin manually, then try running `rocketpool service install -d` again to finish updating."
                echo -e "${RESET}"
                exit 1
            fi

        ;;

    esac
fi


# Check for existing installation
progress 5 "Checking for existing installation..."
if [ -d $RP_PATH ]; then 
    # Check for legacy files - key on the old config.yml
    if [ -f "$RP_PATH/config.yml" ]; then
        progress 5 "Old installation detected, backing it up and migrating to new config system..."
        OLD_CONFIG_BACKUP_PATH="$RP_PATH/old_config_backup"
        { mkdir -p $OLD_CONFIG_BACKUP_PATH || fail "Could not create old config backup folder."; } >&2

        if [ -f "$RP_PATH/config.yml" ]; then 
            { mv "$RP_PATH/config.yml" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move config.yml to backup folder."; } >&2
        fi
        if [ -f "$RP_PATH/settings.yml" ]; then 
            { mv "$RP_PATH/settings.yml" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move settings.yml to backup folder."; } >&2
        fi
        if [ -f "$RP_PATH/docker-compose.yml" ]; then 
            { mv "$RP_PATH/docker-compose.yml" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move docker-compose.yml to backup folder."; } >&2
        fi
        if [ -f "$RP_PATH/docker-compose-metrics.yml" ]; then 
            { mv "$RP_PATH/docker-compose-metrics.yml" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move docker-compose-metrics.yml to backup folder."; } >&2
        fi
        if [ -f "$RP_PATH/docker-compose-fallback.yml" ]; then 
            { mv "$RP_PATH/docker-compose-fallback.yml" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move docker-compose-fallback.yml to backup folder."; } >&2
        fi
        if [ -f "$RP_PATH/prometheus.tmpl" ]; then 
            { mv "$RP_PATH/prometheus.tmpl" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move prometheus.tmpl to backup folder."; } >&2
        fi
        if [ -f "$RP_PATH/grafana-prometheus-datasource.yml" ]; then 
            { mv "$RP_PATH/grafana-prometheus-datasource.yml" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move grafana-prometheus-datasource.yml to backup folder."; } >&2
        fi
        if [ -d "$RP_PATH/chains" ]; then 
            { mv "$RP_PATH/chains" "$OLD_CONFIG_BACKUP_PATH" || fail "Could not move chains directory to backup folder."; } >&2
        fi
    fi

    # Back up existing config file
    if [ -f "$RP_PATH/user-settings.yml" ]; then
        progress 5 "Backing up configuration settings to user-settings-backup.yml..."
        { cp "$RP_PATH/user-settings.yml" "$RP_PATH/user-settings-backup.yml" || fail "Could not backup configuration settings."; } >&2
    fi
fi


# Create ~/.rocketpool dir & files
progress 6 "Creating Rocket Pool user data directory..."
{ mkdir -p "$DATA_PATH/validators" || fail "Could not create the Rocket Pool user data directory."; } >&2
{ mkdir -p "$RP_PATH/runtime" || fail "Could not create the Rocket Pool runtime directory."; } >&2
{ mkdir -p "$DATA_PATH/secrets" || fail "Could not create the Rocket Pool secrets directory."; } >&2
{ mkdir -p "$DATA_PATH/rewards-trees" || fail "Could not create the Rocket Pool rewards trees directory."; } >&2
{ mkdir -p "$RP_PATH/extra-scrape-jobs" || fail "Could not create the Prometheus extra scrape jobs directory."; } >&2
{ mkdir -p "$RP_PATH/alerting/rules" || fail "Could not create the alerting rules directory."; } >&2


# Copy package files
progress 7 "Copying package files to Rocket Pool user data directory..."
{ cp -r "$PACKAGE_FILES_PATH/addons" "$RP_PATH" || fail "Could not copy addons folder to the Rocket Pool user data directory."; } >&2
{ cp -r -n "$PACKAGE_FILES_PATH/override" "$RP_PATH" || rsync -r --ignore-existing "$PACKAGE_FILES_PATH/override" "$RP_PATH" || fail "Could not copy new override files to the Rocket Pool user data directory."; } >&2
{ cp -r "$PACKAGE_FILES_PATH/scripts" "$RP_PATH" || fail "Could not copy scripts folder to the Rocket Pool user data directory."; } >&2
{ cp -r "$PACKAGE_FILES_PATH/templates" "$RP_PATH" || fail "Could not copy templates folder to the Rocket Pool user data directory."; } >&2
{ cp -r "$PACKAGE_FILES_PATH/alerting" "$RP_PATH" || fail "Could not copy alerting folder to the Rocket Pool user data directory."; } >&2
{ cp "$PACKAGE_FILES_PATH/grafana-prometheus-datasource.yml" "$PACKAGE_FILES_PATH/prometheus.tmpl" "$RP_PATH" || fail "Could not copy base files to the Rocket Pool user data directory."; } >&2
{ find "$RP_PATH/scripts" -name "*.sh" -exec chmod +x {} \; 2>/dev/null || fail "Could not set executable permissions on package files."; } >&2
{ touch -a "$RP_PATH/.firstrun" || fail "Could not create the first-run flag file."; } >&2

# Clean up unnecessary files from old installations
progress 8 "Cleaning up obsolete files from previous installs..."
{ rm -rf "$DATA_PATH/fr-default" || echo "NOTE: Could not remove '$DATA_PATH/fr-default' which is no longer needed."; } >&2
GRAFFITI_OWNER=$(stat -c "%U" $RP_PATH/addons/gww/graffiti.txt)
if [ "$GRAFFITI_OWNER" = "$USER" ]; then
    { rm -f "$RP_PATH/addons/gww/graffiti.txt" || echo -e "${COLOR_YELLOW}WARNING: Could not remove '$RP_PATH/addons/gww/graffiti.txt' which was used by the Graffiti Wall Writer addon. You will need to remove this file manually if you intend to use the Graffiti Wall Writer.${COLOR_RESET}"; } >&2
fi
}
# Remove deprecated version tags
find $RP_PATH/override/ -name "*.yml"  -exec sed -i '/^version: "3\.7"$/d' {} \;
find $RP_PATH/templates/ -name "*.tmpl"  -exec sed -i '/^version: "3\.7"$/d' {} \;

install "$@"

