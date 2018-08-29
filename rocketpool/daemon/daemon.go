package daemon

import (
    "io/ioutil"
    "os"
    "os/exec"
)


// Config
const servicePath string = "/lib/systemd/system/rocketpool.service"
const daemonPath string = "/usr/bin/rocketpool"
const checkRPIPVoteInterval string = "10s"


// Install daemon
func Install() error {

    // Service config
    config := []byte(
        "[Unit]" + "\r\n" +
        "Description=Rocket Pool smartnode daemon" + "\r\n" +
        "After=network.target" + "\r\n" +
        "\r\n" +
        "[Service]" + "\r\n" +
        "Type=simple" + "\r\n" +
        "ExecStart=" + daemonPath + " service run" + "\r\n" +
        "Restart=always" + "\r\n" +
        "RestartSec=5" + "\r\n" +
        "\r\n" +
        "[Install]" + "\r\n" +
        "WantedBy=multi-user.target" +
    "\r\n")

    // Write service config to systemd path
    err := ioutil.WriteFile(servicePath, config, 0664)
    if err != nil {
        return err
    }

    // Reload systemd services
    return exec.Command("systemctl", "daemon-reload").Run()

}


// Uninstall daemon
func Uninstall() error {

    // Delete service config at systemd path
    err := os.Remove(servicePath)
    if err != nil {
        return err
    }

    // Reload systemd services
    return exec.Command("systemctl", "daemon-reload").Run()

}


// Enable / disable daemon start at boot
func Enable() error {
    return exec.Command("systemctl", "enable", "rocketpool").Run()
}
func Disable() error {
    return exec.Command("systemctl", "disable", "rocketpool").Run()
}


// Start / stop daemon
func Start() error {
    return exec.Command("systemctl", "start", "rocketpool").Run()
}
func Stop() error {
    return exec.Command("systemctl", "stop", "rocketpool").Run()
}


// Run daemon
func Run() {

    // Check for node exit on minipool removed
    go startCheckNodeExit()
    go checkNodeExit()

    // Check for RPIP alerts on new proposal
    go startCheckRPIPAlerts()
    go checkRPIPAlerts()

    // Check RPIP votes periodically
    go startCheckRPIPVotes(checkRPIPVoteInterval)
    go checkRPIPVotes()

    // Block thread
    select {}

}

