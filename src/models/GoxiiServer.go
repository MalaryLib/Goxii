package models

import (
    "bytes"
    "fmt"
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
    Done bool
}

// A global function that can be used to standardize how we read from 
// connections.
func (g* GoxiiTunnel) Read(buff *bytes.Buffer, conn net.Conn) (error) {
   
    buffer := make([]byte, 1024*100)
	// last_buffer := make([]byte, 1024*100)
    last := 0
    last_rep_counter := 0
    for n, err := conn.Read(buffer); n != 0; {
        if err != nil {
			// most likely a timeout, what do we want to do...
			// we will ignore
        } else if last_rep_counter >= 6 {
            break
        } else if n == last {
            last_rep_counter++
            continue
        }
        last = n
		// last_buffer = buffer
    }
    buff.Write(buffer[:last])
    return nil
}

// The handler for the client (i.e., the connection originating from the user).
func (g *GoxiiTunnel) ClientWorker() {
	// good practice to avoid hanging the connection
    g.Client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	
    counter := 0
    for {
		// this closes the connection when we've noticed that 
		// the dest has not sent us any data.
        if counter > 3 {
			g.Done = true

            g.Client.Close()
            close(g.Result)
            close(g.Payload)
            fmt.Printf("%s %s!\n", color.New(color.FgRed).SprintFunc()("\nClosing connection..."), g.Client.RemoteAddr())
            return
        }
        var buff bytes.Buffer
   
		// read from the client, write that to the channel for the 
		// dest
        g.Read(&buff, g.Client) // will block for duration ^
        g.Payload <- buff.String()

		// wait for the dest to write back to us
        payload := <-g.Result
        payload_bytes := len([]byte(payload))
       
		// if we recieved nothing, don't bother writing
        if (payload_bytes == 0) {
            counter++
            continue
        }

		// sending the payload back to the user
        g.Client.Write([]byte(payload))
        
		// console logging
        fmt.Printf("Input : %s\n", color.New(color.FgCyan).SprintFunc()(buff.String()))
        fmt.Printf("Output: %s\n, (Bytes: %d)", color.New(color.FgCyan).SprintFunc()(payload), payload_bytes)
    }
}

// handler for the destination
func (g *GoxiiTunnel) EndpointWorker() {
    // connect to the service
    conn, err := net.Dial("tcp", g.Destination.Address)
    if err != nil {
        fmt.Printf("%s %s!\n", color.New(color.FgRed).SprintFunc()("Failed to bind to: "), g.Destination.Address)
        g.Client.Close()
        return
    }

	// required for dealing with the keep-alive
	// wait time.
	conn.SetDeadline(time.Now().Add(time.Millisecond * 700))

	// received data from the destination
    for {
        var buff bytes.Buffer
        str := <-g.Payload

        fmt.Fprint(conn, str)
        g.Read(&buff, conn)

		// useful for preventing a potential race condition, but may still lock \_+=+_/
        if (g.Done) {
            conn.Close()
            fmt.Printf("%s %s!\n", color.New(color.FgRed).SprintFunc()("\nClosing connection..."), conn.RemoteAddr())
            return
        }
        g.Result <- buff.String()
    }
}

type GoxiiServer struct {
    Tunnels []GoxiiTunnel
    Endpoint RegisteredEndpoint
    Ls  net.Listener
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
    remote := strings.Split(strings.Split(conn.RemoteAddr().String(), ":")[0], ".")
    for _, elem := range g.Endpoint.Allowed {
        if strings.Contains(elem, ":") {
            elem = strings.Split(elem, ":")[0]
        }

        exitFlag := false
        for index, segment := range strings.Split(elem, ".") {
            if exitFlag {
                break
            } else if (segment == "*") {
                continue
            } else if remote[index] != segment {
                exitFlag = true
                break
            }
        }

        found = !exitFlag
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
    limiter := make(chan int)

    Tunnel := GoxiiTunnel {
        Client: conn,
        Destination: g.Endpoint,
        Payload: payload,
        Result: result,
        Limiter: limiter,
    }
   
    // we nowconn.RemoteAddr().String() start a goroutine that reads
    // from the client and will write to the payload
    // channel.
    go Tunnel.ClientWorker()
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