package main

import (
	"fmt"
	"net"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/gookit/color"
)

const usage = `Print the route packets trace to network host showing AS information

Usage:
  tracert-as <host>

Options:
  -h --help    Show this message`

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "-h")
	}
	arguments, _ := docopt.ParseDoc(usage)
	host, _ := arguments.String("<host>")
	err := traceRoute(host)
	if err != nil {
		color.Red.Println(err)
	}
}

func traceRoute(host string) error {
	ip, err := getIPAddress(host)
	if err != nil {
		return err
	}
	color.Yellow.Println(ip)
	return nil
}

func getIPAddress(host string) (net.IP, error) {
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, fmt.Errorf("%s is invalid", host)
		}
		ip = ips[0]
	}
	return ip, nil
}
