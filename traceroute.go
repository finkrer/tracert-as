package main

import (
	"math/rand"
	"net"
	"time"

	"github.com/gookit/color"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Hop stores information about a single traceroute hop.
type Hop struct {
	Addr    net.Addr
	Rtt     time.Duration
	Success bool
}

// NewHop returns a new Hop value.
func NewHop(addr net.Addr, rtt time.Duration, success bool) Hop {
	return Hop{Addr: addr, Rtt: rtt, Success: success}
}

// TraceRoute returns a channel of hop information values.
func TraceRoute(host string) (err error) {
	dest, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return err
	}

	ttl := 1
	timeout := time.Second

	for {
		color.Normal.Printf("%3d ", ttl)
		hop := sendEcho(dest, ttl, ttl, timeout)
		if !hop.Success {
			color.Red.Printf("%15s\n", "*")
		}
		ttl++
		if hop.Success {
			color.Yellow.Printf("%15s", hop.Addr.String())
			color.Normal.Printf(": ")
			color.Blue.Printf("%9s\n", hop.Rtt.Truncate(time.Microsecond))
			if hop.Addr.String() == dest.String() {
				break
			}
			timeout = hop.Rtt*3 + time.Millisecond*50
		}
		color.Normal.Println()
	}

	return nil
}

func sendEcho(dest net.Addr, seq, ttl int, timeout time.Duration) (hop Hop) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return Hop{Success: false}
	}
	defer conn.Close()

	echo, err := createICMPEcho(seq)
	if err != nil {
		return Hop{Success: false}
	}
	conn.IPv4PacketConn().SetTTL(ttl)

	start := time.Now()

	_, err = conn.WriteTo(echo, dest)
	if err != nil {
		return Hop{Success: false}
	}

	reply := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return Hop{Success: false}
	}
	_, peer, err := conn.ReadFrom(reply)
	if err != nil {
		return Hop{Success: false}
	}

	rtt := time.Since(start)

	return NewHop(peer, rtt, true)
}

func createICMPEcho(seq int) (request []byte, err error) {
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   rand.Int(),
			Seq:  seq,
			Data: []byte(""),
		}}
	return message.Marshal(nil)
}
