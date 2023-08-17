package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	InvalidConnectionHttpFile = "./templates/Whoops.html"
)

type Proxy struct {
	MacAllowedMap map[string]bool
	ConnectionPool *sync.Pool
}

type ProxyObserver struct {

}

func ProxyTimerTriggered(Inactive *bool, TimerActive *bool, multiplier int, unit time.Duration) {
	println("Starting a timer...")
	*Inactive = false
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(w *sync.WaitGroup){
		*TimerActive = true
		time.Sleep(4 * time.Second)
		if *TimerActive {
			*Inactive = true
			*TimerActive = false
		}
		w.Done()
	}(&wg)
	
	wg.Wait()
}

func (p *Proxy) ProxyData(src net.Conn, dst io.Writer, wg *sync.WaitGroup) {
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

func (p *Proxy) ServeHTTPViaConn(HttpFilePath string, conn net.Conn) {
	var dat []byte
	dat, err := os.ReadFile(HttpFilePath)
	if err != nil {
		dat = []byte("You shouldn't be here")
	}

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: %d\r\n\r\n%s", len(dat), string(dat))
	conn.Write([]byte(response))
}

func (p *Proxy) handleConnection(conn net.Conn, DestinationAddress *string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	// we need to verify that the connection is valid based on the 
	// mac address.
	db, ok := p.ConnectionPool.Get().(*DatabaseConn)
	defer p.ConnectionPool.Put(db)
	if !ok || db == nil {
		// there were not enough connections in the pool,
		// so we are going to wait it out.
		return
	}

	mac := db.GetMacFromIP(ip)
	fmt.Printf("Received the MAC: %s\n", mac)
	valid, ok := p.MacAllowedMap[mac]
	if !ok || (ok && !valid) {
		p.ServeHTTPViaConn(InvalidConnectionHttpFile, conn)
		return
	}

	// this is a valid connection so we will proxy the data to the
	// destination server...
	dest, err := net.Dial("tcp", *DestinationAddress)
	if err != nil {
		panic(err)
	}
	
	ConnWg := &sync.WaitGroup{}
	ConnWg.Add(2)
	go p.ProxyData(conn, dest, ConnWg)
	go p.ProxyData(dest, conn, ConnWg)

	ConnWg.Wait()
}

func (p *Proxy) StartProxy(BindPort int, DestinationAddress string, ctx context.Context) {
	ls, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", BindPort))
	check(err)
	defer ls.Close()
	SubtleText("Activating the Proxy!\n")

	wg := sync.WaitGroup{}
	proxy_loop:
	for {
		select {
		case <-ctx.Done():
			break proxy_loop
		default:
			conn, _ := ls.Accept()
			wg.Add(1)
			go p.handleConnection(conn, &DestinationAddress, &wg)
		}
	}

	wg.Wait()
}