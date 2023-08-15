package main

import (
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PacketHandler struct {
	Interface string
	Snaplen int32
	Primsc bool
	Timeout time.Duration
	Filter string
}

func (p *PacketHandler) GetHandle() (handle *pcap.Handle, err error) {
	handle, err = pcap.OpenLive(p.Interface, p.Snaplen, p.Primsc, p.Timeout)
	if err != nil {
		panic(err)
	}

	handle.SetBPFFilter(p.Filter)
	return handle, err
}

func (p *PacketHandler) GetSource(handle *pcap.Handle) (*gopacket.PacketSource) {
	return gopacket.NewPacketSource(handle, handle.LinkType())
}

func ToEthernet(packet gopacket.Packet) (*layers.Ethernet) {
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	ethpack, _ := ethernetLayer.(*layers.Ethernet)

	return ethpack
}

func ToIPv4(packet gopacket.Packet) (*layers.IPv4) {
	IPv4 := packet.Layer(layers.LayerTypeIPv4)
	ip, _ := IPv4.(*layers.IPv4)
	
	return ip
}


