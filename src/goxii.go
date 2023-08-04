package main

import (
	"goxii/models"
	"os"
	"strconv"
	"strings"
)

func main() {
	ListenPort := 8080
	IpListPath := "../resources/.ips"
	var Address string

	// <port to listen> <ip to register:port> 
	args := os.Args[1:]
	if len(args) == 2 {
		ListenPort, _ = strconv.Atoi(args[0])
		Address = args[1]
	}

	// parse allowed ips
	var allowed []string
	// parse the allowed ips into memory
	dat, err := os.ReadFile(IpListPath)
	if err == nil {
		// we want to read the file and split it line by line
		for _, address := range strings.Split(string(dat), "\n") {
			if len(address) > 0 {
				allowed = append(allowed, address)
			}
		}
	}

	g := models.GoxiiServer{}
	g.RegisterEndpoint(Address, allowed)

	g.Init(ListenPort)
	g.Start()
}