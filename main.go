package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

// configuration file
type Configuration struct {

	// packet barrier interface
	PacketInterface string

	// related to this proxy
	HostPort int
	DestinationAddress string

	// Proxy Chaining Related Settings
	ProxyDescendents []string

	// related to this this MacServer
	InheritMacServer bool
	MacServerAddress string
	MacServerPort int

	// Database Related Fields
	PQUser string
	PQPassword string
	PQHost string
	DatabaseName string

	// Table(s)
	IpToMacTableName string
}

func pullConfiguration(filePath string, IsDescendent bool) (*Configuration) {
	if IsDescendent {
		return &Configuration{
			PacketInterface: "lo",
			HostPort: 8081,
		}
	}
	file, _ := os.Open(filePath)
	defer file.Close()

	decoder := json.NewDecoder(file)
	configuration := &Configuration{}
	decoder.Decode(configuration)

	return configuration
}

// full command used for running goxii
// Goxii --port 8080 --destination 127.0.0.1:8081 --mac
func main() {
	// parse our argument flags
	ConfigurationPath := flag.String("conf", "./configuration.json", "The path to the config file for this proxy!")
	IsDescendant := flag.Bool("descendent", false, "Set to true if this proxy is a child of another proxy!")
	flag.Parse()

	// set up the conf struct
	conf := pullConfiguration(*ConfigurationPath, *IsDescendant)
	SubtleTextIndent(fmt.Sprintf("Proxy bound to :%d\n", conf.HostPort), false)
	SubtleTextIndent(fmt.Sprintf("Proxy Target: %s (via %s)\n", conf.DestinationAddress, conf.ProxyDescendents), false)
	SubtleTextIndent(fmt.Sprintf("Mac Ingestion Server: %s\n", conf.MacServerAddress), false)

	// creates the global context for closing shop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// the signal channel
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT)

	// Primitives needed for both descendents and 
	// parent proxies.
	Connection := &DatabaseConn{
		User: conf.PQUser,
		Password: conf.PQPassword,
		Database: conf.DatabaseName,
		Host: conf.PQHost,
		TableName: conf.IpToMacTableName,
	}
	Connection.InitConnection(conf.PQUser, conf.PQPassword, conf.PQHost, conf.DatabaseName, conf.IpToMacTableName)

	// instantiating our maps
	MacAllowedMap := make(map[string]bool, 0)

	if *IsDescendant {
		var proxy_ip string
		// we do not need to do a configuration pull, but we do need to know who our dest 
		// is, so we wait to receive an http request with a specific code.
		WarningText(fmt.Sprintf("Started the descendent server on port %d...\n", conf.HostPort))
		ds := DescendentSetupServer{}
		ds.Listen(conf.HostPort, &proxy_ip)

		// received the ip that we will use in the destination address spot
		// so we just do some automagic
		conf.DestinationAddress = proxy_ip
		fmt.Printf("Received Proxy Dest: %s", proxy_ip)
	} else {
		WarningText(fmt.Sprintf("Packet Barrier Interface: %s\n", conf.PacketInterface))

		// packet barrier set-up
		pb := InitPacketBarrier(conf.PacketInterface)
		pb.Connection = Connection
		go pb.StartPacketBarrier(ctx)

		// starting a local pb
		pbl := InitPacketBarrier("lo")
		pbl.Connection = Connection
		go pbl.StartPacketBarrier(ctx)


		// starting our mac server
		ms := MacServer{
			MacAllowedMap: MacAllowedMap,
			ProxyDescendents: conf.ProxyDescendents,
			Dp: DestinationPayload{
				DestinationAddress: conf.DestinationAddress,
			},
		}
		go ms.StartServer(conf.MacServerAddress, conf.MacServerPort)
		
	}

	// starting the proxy service
	Proxy := Proxy{
		Connection: Connection,
		MacAllowedMap: MacAllowedMap,
		IsDescendant: *IsDescendant,
		ProxyChain: conf.ProxyDescendents,
	}
	go Proxy.StartProxy(conf.HostPort, conf.DestinationAddress, ctx)
	
	H1Print("All services are online...\n")
	<- exit
	WarningText("\n\nReceived SIGINT, waiting for other services to return...\n")

	cancel()
}