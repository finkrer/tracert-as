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
	Type    icmp.Type
	Success bool
}

// TraceRoute returns a channel of hop information values.
func TraceRoute(host string) (<-chan Hop, <-chan error) {
	errc := make(chan error, 1)

	dest, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		errc <- fmt.Errorf("%s is invalid", host)
		defer close(errc)
		return nil, errc
	}

	ttl := 1
	timeout := time.Second

	out := make(chan Hop)
	go func() {
		defer close(out)
		defer close(errc)
		for {
			hop, err := sendEcho(dest, ttl, ttl, timeout)
			if err != nil {
				errc <- err
				break
			}
			out <- hop
			ttl++
			if hop.Success {
				if hop.Type == ipv4.ICMPTypeEchoReply {
					break
				}
				timeout = hop.Rtt*3 + time.Millisecond*50
			}
			if ttl > maxHops {
				return
			}
		}
	}()

	return out, errc
}

func sendEcho(dest net.Addr, seq, ttl int, timeout time.Duration) (hop Hop, err error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return Hop{}, err
	}
	defer conn.Close()

	echo, err := createICMPEcho(seq)
	if err != nil {
		return Hop{}, err
	}
	conn.IPv4PacketConn().SetTTL(ttl)

	start := time.Now()

	_, err = conn.WriteTo(echo, dest)
	if err != nil {
		return Hop{}, err
	}

	reply := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return Hop{}, err
	}
	_, peer, err := conn.ReadFrom(reply)
	if err != nil {
		return Hop{Number: ttl, Success: false}, nil
	}

	rtt := time.Since(start)

	message, err := icmp.ParseMessage(1, reply)
	if err != nil {
		return Hop{Number: ttl, Success: false}, nil
	}

	return Hop{Number: ttl, Addr: peer, Rtt: rtt, Type: message.Type, Success: true}, nil
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
