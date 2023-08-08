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

func (g *GoxiiTunnel) SendEofReq(conn net.Conn) {
    fmt.Fprint(conn, "HTTP/1.1 102 Processing\r\n\r\n")
}

func (g* GoxiiTunnel) Read(buff *bytes.Buffer, conn net.Conn) (error) {
   
    buffer := make([]byte, 1024*100)
    last := 0
    last_rep_counter := 0
    for n, err := conn.Read(buffer); n != 0; {
        if err != nil {
            panic(err)
        } else if last_rep_counter == 3 {
            break
        } else if n == last {
            last_rep_counter++
            continue
        }
        last = n
    }
    buff.Write(buffer[:last])
    return nil
}

func (g *GoxiiTunnel) ClientWorker() {
    g.Client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
    counter := 0
    for {
        if counter > 3 {
            g.Client.Close()
            close(g.Result)
            close(g.Payload)
            g.Done = true
            fmt.Printf("%s %s!\n", color.New(color.FgRed).SprintFunc()("\nClosing connection..."), g.Client.RemoteAddr())
            return
        }
        var buff bytes.Buffer
   
        g.Read(&buff, g.Client) // will block for duration ^
        g.Payload <- buff.String()

        payload := <-g.Result
        payload_bytes := len([]byte(payload))
       
        if (payload_bytes == 0) {
            counter++
            continue
        }

        g.Client.Write([]byte(payload))
        // some house keeping

        fmt.Printf("Input : %s\n", color.New(color.FgCyan).SprintFunc()(buff.String()))
        fmt.Printf("Output: %s\n, (Bytes: %d)", color.New(color.FgCyan).SprintFunc()(payload), payload_bytes)
    }
}

func (g *GoxiiTunnel) EndpointWorker() {
    // connect to the service
    conn, err := net.Dial("tcp", g.Destination.Address)
    if err != nil {
        fmt.Printf("%s %s!\n", color.New(color.FgRed).SprintFunc()("Failed to bind to: "), g.Destination.Address)
        g.Client.Close()
        return
    }

    for {
       
        var buff bytes.Buffer
        str := <-g.Payload

        fmt.Fprint(conn, str)
        g.Read(&buff, conn)
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