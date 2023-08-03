package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerSocket struct {
	ls net.Listener
	allowed []string
	dest net.Conn
	client net.Conn
	destination_address string
	destination_port int
}

func (s *ServerSocket) init(server_port int, destination_address string, destination_port int, allowed_list_fp string) {
	// init the server
	server_address := fmt.Sprintf("0.0.0.0:%d", server_port)
	ls, err := net.Listen("tcp", server_address)
	if err != nil {
		panic(err)
	}

	// init the dest conn
	dest, err := net.Dial("tcp", fmt.Sprintf("%s:%d", destination_address, destination_port))
	if err != nil {
		panic(err)
	}

	s.dest = dest
	fmt.Printf("[+] Connected to %s:%d\n", destination_address, destination_port)
	
	s.ls = ls

	// parse the allowed ips into memory
	dat, err := os.ReadFile(allowed_list_fp)
	if err == nil {
		// we want to read the file and split it line by line
		for _, address := range strings.Split(string(dat), "\n") {
			if len(address) > 0 {
				s.allowed = append(s.allowed, address)
			}
		}
	}

	// allowing the following ips output
	fmt.Println("Allowing the following IPs:")
	for _, address := range s.allowed {
		fmt.Println(address)
	}

	s.destination_address = destination_address
	s.destination_port = destination_port
}

func (s *ServerSocket) manage(conn net.Conn) {
	// we need to take the input from conn and write it to 
	// the output of the server that we are going to write to.
	var cBuff bytes.Buffer
	var dBuff bytes.Buffer

	dest, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.destination_address, s.destination_port))
	if err != nil {
		panic(err)
	}

	s.dest = dest

	// var tBuff bytes.Buffer
	// var dBuff bytes.Buffer
	// info := make(chan string)

	// start a goroutine for cBuff events
	go func() {
		fmt.Println("Starting to manage a client...")
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

		_, err := io.Copy(&cBuff, conn)
		if err != io.EOF && err != nil {
			fmt.Fprint(conn, "HTTP/1.1 102 Processing\r\n\r\n")
		}

		str, _ := io.ReadAll(&cBuff)
		println(string(str))		
		fmt.Fprint(s.dest, string(str))
		fmt.Fprint(s.dest, string(str))
		io.Copy(&dBuff, s.dest)
		response, _ := io.ReadAll(&dBuff)
		fmt.Fprint(conn, string(response))
		cBuff.Reset()
		dBuff.Reset()
		conn.Close()
		s.dest.Close()
	}()
}

func (s *ServerSocket) handle(conn net.Conn, remote string) {
	var found bool = false
	
	for _, element := range s.allowed {
		if (element == remote) {
			found = true
			break
		}
	}

	if !found {
		fmt.Println("Blocked yo ass! (", remote, ")")
		conn.Close()
		return
	}

	fmt.Println("Let em thru!")
	go s.manage(conn)
}

func (s *ServerSocket) listen() {
	for {
		conn, err := s.ls.Accept()
		if err != nil {
			panic(err)
		}

		// perform some validation based on the known inputs
		remoteAddr := strings.Split(conn.RemoteAddr().String(), ":")
		s.client = conn
		go s.handle(s.client, remoteAddr[0])
	}
}

func main() {
	// initialize all required arguments
	args := os.Args[1:]
	if len(args) != 4 {
		return;
	}

	// match the args with their variables
	server_port, err := strconv.Atoi(args[0])
	destination_address := args[1]
	destination_port, _ := strconv.Atoi(args[2])
	allowed_ip_fpath := args[3]

	if err != nil {
		panic(err)
	}

	// now we begin listening
	socket := ServerSocket{}
	socket.init(server_port, destination_address, destination_port, allowed_ip_fpath)
	socket.listen()

	fmt.Printf("Goxii [0.0.0.0:%d] -> %s:%d\n", server_port, destination_address, destination_port)
}