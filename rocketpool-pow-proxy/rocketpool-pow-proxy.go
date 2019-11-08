package main

import (
    "github.com/rocket-pool/smartnode/rocketpool-pow-proxy/proxy"
)


// Run application
func main() {

    // Initialise and start POW proxy server
    proxyServer := proxy.NewProxyServer()
    proxyServer.Start()

}

