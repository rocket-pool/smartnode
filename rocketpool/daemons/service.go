package daemons

import (
    "io/ioutil"
    "os"
    "os/exec"
)


// Config
const SERVICE_PATH string = "/lib/systemd/system/"
const DAEMON_PATH string = "/usr/bin/rocketpool"


// Install daemon
func Install(name string) error {

    // Service config
    config := []byte(
        "[Unit]" + "\r\n" +
        "Description=Rocket Pool " + name + " daemon" + "\r\n" +
        "After=network.target" + "\r\n" +
        "\r\n" +
        "[Service]" + "\r\n" +
        "Type=simple" + "\r\n" +
        "ExecStart=" + DAEMON_PATH + " service " + name + " run" + "\r\n" +
        "Restart=always" + "\r\n" +
        "RestartSec=5" + "\r\n" +
        "\r\n" +
        "[Install]" + "\r\n" +
        "WantedBy=multi-user.target" +
    "\r\n")

    // Write service config to systemd path
    if err := ioutil.WriteFile(SERVICE_PATH + "rocketpool-" + name + ".service", config, 0664); err != nil {
        return err
    }

    // Reload systemd services
    return exec.Command("systemctl", "daemon-reload").Run()

}


// Uninstall daemon
func Uninstall(name string) error {

    // Delete service config at systemd path
    if err := os.Remove(SERVICE_PATH + "rocketpool-" + name + ".service"); err != nil {
        return err
    }

    // Reload systemd services
    return exec.Command("systemctl", "daemon-reload").Run()

}


// Enable / disable daemon start at boot
func Enable(name string) error {
    return exec.Command("systemctl", "enable", "rocketpool-" + name).Run()
}
func Disable(name string) error {
    return exec.Command("systemctl", "disable", "rocketpool-" + name).Run()
}


// Start / stop daemon
func Start(name string) error {
    return exec.Command("systemctl", "start", "rocketpool-" + name).Run()
}
func Stop(name string) error {
    return exec.Command("systemctl", "stop", "rocketpool-" + name).Run()
}


// Get daemon status
func Status(name string) string {
    out, _ := exec.Command("systemctl", "status", "rocketpool-" + name).Output()
    return string(out[:])
}

