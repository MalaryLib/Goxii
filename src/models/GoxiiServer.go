package models

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
	"github.com/fatih/color"
)

type RegisteredEndpoint struct {
	Address string
	Allowed []string
}

type GoxiiTunnel struct {
	Client net.Conn
	Destination RegisteredEndpoint
	Payload (chan string)
	Result (chan string)
	Limiter (chan int)
}

func (g *GoxiiTunnel) SendEofReq(conn net.Conn) {
	fmt.Fprint(conn, "HTTP/1.1 102 Processing\r\n\r\n")
}

func (g* GoxiiTunnel) Read(buff *bytes.Buffer, conn net.Conn) (error) {
	_, err := io.Copy(buff, conn)
	if err != io.EOF && err != nil {
		// this is the timeout, request is assumed to be done.
		// this is not my best decision, but \_(0-0)_/
		g.SendEofReq(conn)
		return err
	}
	return err
}

func (g *GoxiiTunnel) ClientWorker(conn net.Conn) {
	g.Client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var buff bytes.Buffer
	
	g.Read(&buff, conn)	// will block for duration ^
	fmt.Printf("%s\n", color.New(color.FgCyan).SprintFunc()(buff.String()))

	g.Payload <- buff.String()

	g.Client.Write([]byte(<-g.Result))
	fmt.Printf("Was written To: %s\n", color.New(color.FgCyan).SprintFunc()())

	// some house keeping
	g.Client.Close()
	close(g.Result)
	close(g.Payload)
}

func (g *GoxiiTunnel) EndpointWorker() {
	// connect to the service
	conn, err := net.Dial("tcp", g.Destination.Address)
	if err != nil {
		fmt.Printf("%s %s!\n", color.New(color.FgRed).SprintFunc()("Failed to bind to: "), g.Destination.Address)
		g.Client.Close()
		return
	}
	var buff bytes.Buffer
	str := <-g.Payload
	fmt.Printf("To Endpoint: %s\n", color.New(color.FgCyan).SprintFunc()(str))

	fmt.Fprint(conn, str)
	g.Read(&buff, conn)
	fmt.Printf("Endpoint Sent: %s\n", color.New(color.FgCyan).SprintFunc()(buff.String()))

	
	g.Result <- buff.String()
	conn.Close()
}

type GoxiiServer struct {
	Tunnels []GoxiiTunnel
	Endpoint RegisteredEndpoint
	Ls 	net.Listener
}

func (g *GoxiiServer) RegisterEndpoint(Address string, Allowed []string) {
	Endpoint := RegisteredEndpoint {
		Address: Address,
		Allowed: Allowed,
	}
	
	g.Endpoint = Endpoint
}

func (g *GoxiiServer) Init(port int) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}

	ls, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	// reassigning to our struct
	g.Ls = ls
}

func (g *GoxiiServer) Verify(conn net.Conn) (found bool) {
	found = false
	remote := strings.Split(conn.RemoteAddr().String(), ":")[0]
	for _, elem := range g.Endpoint.Allowed {
		if elem == remote {
			found = true
		}
	}

	if !found {
		conn.Close()
	}

	return found
}

func (g *GoxiiServer) StartTunnel(conn net.Conn) {
	// this function will deal with the tunneling part of the 
	// server.
	payload := make(chan string)
	result := make(chan string)
	limiter := make(chan int, 2)

	Tunnel := GoxiiTunnel {
		Client: conn,
		Destination: g.Endpoint,
		Payload: payload,
		Result: result,
		Limiter: limiter,
	}

	g.Tunnels = append(g.Tunnels, Tunnel)
	
	// we nowconn.RemoteAddr().String() start a goroutine that reads
	// from the client and will write to the payload
	// channel.
	go Tunnel.ClientWorker(conn)
	go Tunnel.EndpointWorker()
}

func (g *GoxiiServer) Start() {
	fmt.Printf("Starting to listen on %s\n", color.New(color.FgHiYellow).SprintFunc()(g.Ls.Addr().String()))

	for {
		conn, err := g.Ls.Accept()
		if err != nil {
			panic(err)
		}

		if g.Verify(conn) {
			fmt.Printf("(Allowed) %s\n", color.New(color.FgHiGreen).SprintFunc()(conn.RemoteAddr().String()))
			go g.StartTunnel(conn) 
		} else {
			fmt.Printf("(Rejected) %s\n", color.New(color.FgHiRed).SprintFunc()(conn.RemoteAddr().String()))
		}
		
	}
}