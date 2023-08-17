package main

import (
	"context"
	"sync"
	"time"

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
func (p *PacketBarrier) handlePacket(packet gopacket.Packet, c chan <- *IpMacMapping, wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	ethPacket := ToEthernet(packet)
	mac := ethPacket.SrcMAC.String()
	
	ipv4Packet := ToIPv4(packet)
	ip := ipv4Packet.SrcIP.String()

	loop:
	for {
		select {
		case <- ctx.Done():
			break
		default:
			defer func() {
				if r := recover(); r != nil {

				}
			}()
			case c <- &IpMacMapping{ IP: ip, Mac: mac}:
				break loop
			case <- time.After(100 * time.Millisecond):
				// timeout occured on the channel send
				break
		}
	}

	wg.Done()
}

func (p *PacketBarrier) InitMappingService(c <- chan *IpMacMapping, wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	outerloop:
	for {
		select {
		case mapping := <-c:
			db, ok := p.ConnectionPool.Get().(*DatabaseConn)
			defer p.ConnectionPool.Put(db)

			if !ok {
				break
			}

			if db != nil {
				db.InsertMapping(mapping.IP, mapping.Mac)
			}
		case <- ctx.Done():
			break outerloop
		}
	}

	wg.Done()
	p.ClosePacketBarrier = true
}

func (p *PacketBarrier) StartPacketBarrier(ctx context.Context) {
	SubtleText("Activating the packet barrier!\n")

	// set up the context to handle proper exits
	wg := sync.WaitGroup{}

	handle, err := p.Handler.GetHandle()
	check(err)

	source := p.Handler.GetSource(handle)
	MappingChannel := make(chan *IpMacMapping, 10)
	PacketChannel := make(chan gopacket.Packet, 5)

	defer close(MappingChannel)
	defer close(PacketChannel)
	defer handle.Close()

	go p.InitMappingService(MappingChannel, &wg, ctx)
	MainDriver:
	for {
		select {
		case packet := <- source.Packets():
			PacketChannel <- packet
		case pack := <- PacketChannel:
			go p.handlePacket(pack, MappingChannel, &wg, ctx)
		case <- ctx.Done():
			break MainDriver
		}
	}
	WarningText("Closing the Packet Barrier...\n")
	wg.Wait() // waits for all the handlePacket goroutines to exit.
}

