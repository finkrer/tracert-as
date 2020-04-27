package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const maxHops = 64

// Hop stores information about a single traceroute hop.
type Hop struct {
	Number  int
	Addr    net.Addr
	Rtt     time.Duration
	Success bool
}

// NewHop returns a new Hop value.
func NewHop(number int, addr net.Addr, rtt time.Duration, success bool) Hop {
	return Hop{Number: number, Addr: addr, Rtt: rtt, Success: success}
}

// TraceRoute returns a channel of hop information values.
func TraceRoute(host string) (<-chan Hop, error) {
	dest, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, fmt.Errorf("%s is invalid", host)
	}

	ttl := 1
	timeout := time.Second

	out := make(chan Hop)
	go func() {
		defer close(out)
		for {
			hop := sendEcho(dest, ttl, ttl, timeout)
			out <- hop
			ttl++
			if hop.Success {
				if hop.Addr.String() == dest.String() {
					break
				}
				timeout = hop.Rtt*3 + time.Millisecond*50
			}
			if ttl > maxHops {
				return
			}
		}
	}()

	return out, nil
}

func sendEcho(dest net.Addr, seq, ttl int, timeout time.Duration) (hop Hop) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return Hop{Number: ttl, Success: false}
	}
	defer conn.Close()

	echo, err := createICMPEcho(seq)
	if err != nil {
		return Hop{Number: ttl, Success: false}
	}
	conn.IPv4PacketConn().SetTTL(ttl)

	start := time.Now()

	_, err = conn.WriteTo(echo, dest)
	if err != nil {
		return Hop{Number: ttl, Success: false}
	}

	reply := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return Hop{Number: ttl, Success: false}
	}
	_, peer, err := conn.ReadFrom(reply)
	if err != nil {
		return Hop{Number: ttl, Success: false}
	}

	rtt := time.Since(start)

	return NewHop(ttl, peer, rtt, true)
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
