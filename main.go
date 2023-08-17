package main

import (
	// "flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"flag"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

)

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

type ProxyObserver struct {

}

func (p *ProxyObserver) Write(b []byte) (int, error) {
	println(string(b))
	return len(b), nil
}

func StartPacketListener(iface string, HostPort int, IpMacMap map[string]string) {
	handle, err := pcap.OpenLive(iface, 1600, false, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	err = handle.SetBPFFilter("ip")
	if err != nil {
		panic(err)
	}

	defer handle.Close()
	source := gopacket.NewPacketSource(handle, handle.LinkType())
	
	println("Activating the packet barrier!")
	for packet := range source.Packets() {
		// this will listen to packets and provide
		// a mapping between ip and mac address for the
		// server to use.

		eth := ToEthernet(packet)
		ip4 := ToIPv4(packet)

		ipSrc := ip4.SrcIP.String()
		_, ok := IpMacMap[ipSrc]
		if !ok {
			IpMacMap[ipSrc] = strings.ToUpper(eth.SrcMAC.String())
		}
	}
}

func ProxyTimerTriggered(btn *bool, ctx *bool, multiplier int, unit time.Duration) {
	println("Starting a timer...")
	*btn = false
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(w *sync.WaitGroup){
		*ctx = true
		time.Sleep(4 * time.Second)
		if *ctx {
			*btn = true
			*ctx = false
		}
		w.Done()
	}(&wg)
	
	wg.Wait()
}

func ProxyConnHandle(src net.Conn, dst io.Writer, wg *sync.WaitGroup) {
	IsInactive := false
	TimerActive := false
	for {
		src.SetDeadline(time.Now().Add(15 * time.Second))

		n, _ := io.Copy(dst, src)
		if IsInactive {
			println("Closing shop...")
			break
		} else if n == 0 && !TimerActive {
			ProxyTimerTriggered(&IsInactive, &TimerActive, 4, time.Second)
		}
	}
	wg.Done()
}

func StartProxy(
	port int,
	destination string,
	MacAllowedMap map[string]bool,
	ConnectionPool *sync.Pool,
) {
	ls, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}

	defer ls.Close()
	fmt.Printf("Starting to listen on port %d\n", port)

	dat, _ := os.ReadFile("./templates/Whoops.html")
	
	for {
		conn, err := ls.Accept()
		if err != nil {
			panic(err)
		}
		exit := false
		IpSrc := strings.Split(conn.RemoteAddr().String(), ":")[0]
		fmt.Printf("Proxy Request Origin: %s\n", IpSrc)

		db, ok := ConnectionPool.Get().(*DatabaseConn)
		if !ok {
			println("No open connections in the pool!")
			conn.Close()
			ConnectionPool.Put(db)
			continue
		}

		var mac string
		rows, err := db.conn.Query("SELECT * FROM ipmapping WHERE ip = $1", IpSrc)
		for rows.Next() {
			var id int
			var ip, m string
			rows.Scan(&id, &ip, &m)
			mac = m
			break
		}
		ConnectionPool.Put(db)
		
		// perform the mac address verification
		if ok {
			macAllowed, exists := MacAllowedMap[mac]
			if !exists || (exists && !macAllowed) {
				exit = true
			}
		} else {
			exit = true
		}

		// exit if required by mac address verification
		if exit {
			println("Exiting!")
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: %d\r\n\r\n%s", len(dat), string(dat))
			conn.Write([]byte(response))
			conn.Close()
			continue
		}


		// observer := &ProxyObserver{}
		go func(c net.Conn, t string) {
			wg := sync.WaitGroup{}
			wg.Add(2)

			// instantiate a connection to our destination
			// server
			dest, err := net.Dial("tcp", t)
			if err != nil {
				panic(err)
			}
	
			left := io.MultiWriter(conn)
			right := io.MultiWriter(dest) 
	
			// the main proxy behavior
			go ProxyConnHandle(conn, right, &wg)
			go ProxyConnHandle(dest, left, &wg)
			
			wg.Wait()
			dest.Close()
			conn.Close()

			println("Closed a session!")
		}(conn, destination)
	}
	
}

// full command used for running goxii
// Goxii --port 8080 --destination 127.0.0.1:8081 --mac
func main() {
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

	pb := InitPacketBarrier("lo")
	pb.ConnectionPool = ConnectionPool
	
	go pb.StartPacketBarrier()

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

	StartProxy(*HostPortFlag, *DestinationFlag, MacAllowedMap, &sync.Pool{
		New: func() any {
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
	})
	
}