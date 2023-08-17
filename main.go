package main

import (
	// "flag"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}


// full command used for running goxii
// Goxii --port 8080 --destination 127.0.0.1:8081 --mac
func main() {
	// creates the global context for closing shop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// the signal channel
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT)

	// Creates the conneciton pool for the database(s)
	ConnectionPool := &sync.Pool{}
	for i := 0; i < 5; i++ {
		db := &DatabaseConn{
			User: "dev",
			Password: "dev",
			Host: "localhost",
			Database: "ipmacdb",
			TableName: "ipmapping",
		}
		db.InitConnection(db.User, db.Password, db.Host, db.Database, db.TableName)
		ConnectionPool.Put(db)
	}

	Proxy := Proxy{}

	pb := InitPacketBarrier("lo")
	pb.ConnectionPool = ConnectionPool
	
	go pb.StartPacketBarrier(ctx)

	// parse the command line arguments into their variables
	// listed below.
	DestinationFlag := flag.String("destination", "", "Destination IP to connect this proxy to; in the ip:port syntax.")
	HostPortFlag := flag.Int("port", 8901, "Port to bind the proxy on, from the host.")
	// MacFileFlag := flag.String("mac", "./.AllowedMacs", "The path to a file containing a list of hard-coded allowed MAC Addresses.")
	// PacketInterfaceFlag := flag.String("interface", "lo", "The interface we should listen on for packets.")

	flag.Parse()

	// instantiating our maps
	MacAllowedMap := make(map[string]bool, 0)

	// starting our services

	MacIngestionPoint := MacIngestionPoint{
		MacAllowedMap: MacAllowedMap,
	}
	go MacIngestionPoint.StartServer()
	
	Proxy.ConnectionPool = ConnectionPool
	Proxy.MacAllowedMap = make(map[string]bool)
	go Proxy.StartProxy(*HostPortFlag, *DestinationFlag, ctx)
	
	<- exit
	cancel()
}