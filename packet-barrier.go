package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// the packet barrier is the first point of contact that the server has with
// an outside connection and is where we can do our Mac, IP, filtering.
type PacketBarrier struct {
	Handler PacketHandler
	ConnectionPool *sync.Pool
	ClosePacketBarrier bool
}

type IpMacMapping struct {
	IP string
	Mac string
}

func InitPacketBarrier(iface string) *PacketBarrier {

	return &PacketBarrier{
		Handler: PacketHandler{
			Interface: iface,
			Snaplen: 1600,
			Primsc: false,
			Timeout: pcap.BlockForever,
			Filter: "ip",
		},
		ClosePacketBarrier: false,
	}
}

// this function will take the packet and parse it for the proper fields
// such as mac address and src IPs. This will then discard the packet and
// pass it to a ipToMacWorker 
func (p *PacketBarrier) handlePacket(packet gopacket.Packet, c chan <- *IpMacMapping, wg *sync.WaitGroup) {
	ethPacket := ToEthernet(packet)
	mac := ethPacket.SrcMAC.String()
	
	ipv4Packet := ToIPv4(packet)
	ip := ipv4Packet.SrcIP.String()

	c <- &IpMacMapping{
		IP: ip,
		Mac: mac,
	}

	wg.Done()
}

func (p *PacketBarrier) InitMappingService(c <- chan *IpMacMapping, exit chan os.Signal) {
	outerloop:
	for {
		select {
		case mapping := <-c:
			db, ok := p.ConnectionPool.Get().(*DatabaseConn)
			if !ok {
				break
			}

			if db != nil {
				db.InsertMapping(mapping.IP, mapping.Mac)
			}
			p.ConnectionPool.Put(db)
		case <- exit:
			break outerloop
		}
	}

	p.ClosePacketBarrier = true
}

func (p *PacketBarrier) StartPacketBarrier() {
	println("Activating the packet barrier!")
	wg := sync.WaitGroup{}
	MappingServiceBreaker := make(chan os.Signal, 1)
	defer close(MappingServiceBreaker)

	signal.Notify(MappingServiceBreaker, syscall.SIGINT)
	handle, err := p.Handler.GetHandle()

	check(err)
	defer handle.Close()

	source := p.Handler.GetSource(handle)
	MappingChannel := make(chan *IpMacMapping, 10)
	defer close(MappingChannel)

	go p.InitMappingService(MappingChannel, MappingServiceBreaker)
	for packet := range source.Packets() {
		if p.ClosePacketBarrier {
			println("Waiting for all go-routines to exit!")
			wg.Wait()
			break
		}
		wg.Add(1)
		go p.handlePacket(packet, MappingChannel, &wg)
	}
}

