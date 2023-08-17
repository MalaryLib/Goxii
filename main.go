package main

import (
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
	ConnectionPool := &sync.Pool{
		New: func() interface{} {
			db := &DatabaseConn{
				User: "dev",
				Password: "dev",
				Host: "localhost",
				Database: "ipmacdb",
				TableName: "ipmapping",
			}
			db.InitConnection(db.User, db.Password, db.Host, db.Database, db.TableName)
			return db
		},
	}
	for i := 0; i < 9; i++ {
		db, ok := ConnectionPool.New().(*DatabaseConn)
		if !ok {
			panic("This connection pool is acting wierd!")
		}
		ConnectionPool.Put(db)
	}

	pb := InitPacketBarrier("lo")
	pb.ConnectionPool = ConnectionPool
	
	go pb.StartPacketBarrier(ctx)

	// parse the command line arguments into their variables
	// listed below.
	DestinationFlag := flag.String("destination", "", "Destination IP to connect this proxy to; in the ip:port syntax.")
	HostPortFlag := flag.Int("port", 8901, "Port to bind the proxy on, from the host.")
	MacServerFlag := flag.Int("mac-server", 8083, "The port to bind the mac ingestion server to.")
	flag.Parse()

	// instantiating our maps
	MacAllowedMap := make(map[string]bool, 0)

	// starting our services
	Proxy := Proxy{}

	ms := MacServer{
		MacAllowedMap: MacAllowedMap,
	}
	go ms.StartServer(*MacServerFlag)
	
	Proxy.ConnectionPool = ConnectionPool
	Proxy.MacAllowedMap = MacAllowedMap
	go Proxy.StartProxy(*HostPortFlag, *DestinationFlag, ctx)
	
	H1Print("All services are online...\n")
	<- exit
	WarningText("\n\nReceived SIGINT, waiting for other services to return...\n")

	cancel()
}