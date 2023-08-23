package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	
)

type HostProxySettings struct {
	BindPort int
	BindAddress string
}

type DestinationProxySettings struct {
	DestinationEndpoints []string
}

type DatabaseProxySettings struct {
	MacAddressDatabaseUrl string
}

type LoadManagementProxySettings struct {
	MaxConnectionLimit int
	MaxConnectionIdleTime int
}

type ProxySettings struct {
	Host HostProxySettings
	Destination DestinationProxySettings
	Database DatabaseProxySettings
	LoadMgmt LoadManagementProxySettings
}

type SyncProxyItems struct {
	GlobalContext context.Context
	ConnectionWg sync.WaitGroup
}

type Proxy struct {
	ConfigurationServerUrl string
	Settings ProxySettings
	SyncItems SyncProxyItems
}

// Reaches out to the configuration server for information on how to proxy the data
// between the endpoints. Will return an address for a proxy that has already been
//  started with a go-routine. 
func InitializeProxy(ConfigServerUrl string) (*Proxy) {
	p := &Proxy{
		ConfigurationServerUrl: ConfigServerUrl,
	}

	go p.startProxy()

	return p
}

func (p *Proxy) QueryHostProxySettings() {

}

// handler func that connects to the destination endpoint and begins
// the bi-directional communication flow. 
func (p *Proxy) handleConnection(src net.Conn) {
	p.SyncItems.ConnectionWg.Add(1)

	// begin proxying data...
}

// starts the proxy by listening on the configured port.
func (p *Proxy) startProxy() {
	address := fmt.Sprintf("%s:%d", p.Settings.Host.BindAddress, p.Settings.Host.BindPort)
	ls, err := net.Listen("tcp", address)
	if err != nil {
		// there was an error starting the server, we will panic
		panic(err)
	}

	// infinite loop with context checking
	for {
		select {
		case <- p.SyncItems.GlobalContext.Done():
			// begin proxy closure procedure
			break
		default:
			// get a connection for our listener and then goroutine it!
			conn, err := ls.Accept()
			if err != nil {
				log.Panicln(err)
				continue
			}

			go p.handleConnection(conn)
		}
	}

}