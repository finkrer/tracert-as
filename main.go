package main

import (
	"os"
	"time"

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

	hops, err := TraceRoute(host)
	if err != nil {
		color.Red.Println(err)
		return
	}
	for hop := range hops {
		hop.printHop()
	}
}

func (hop *Hop) printHop() {
	color.Normal.Printf("%3d ", hop.Number)
	if !hop.Success {
		color.Red.Printf("%15s\n", "*")
	}
	if hop.Success {
		color.Yellow.Printf("%15s", hop.Addr.String())
		color.Normal.Printf(": ")
		color.Blue.Printf("%9s\n", hop.Rtt.Truncate(time.Microsecond))
	}
	color.Normal.Println()
}
