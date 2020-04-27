package main

import (
	"os"
	"strings"
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

	fi, _ := os.Stdout.Stat()
	if fi.Mode()&os.ModeCharDevice == 0 {
		color.Disable()
	}

	hops, errs := TraceRoute(host)
	for {
		select {
		case err, ok := <-errs:
			if !ok {
				return
			}
			printErr(err)
		case hop, ok := <-hops:
			if !ok {
				return
			}
			printHop(hop)
			if hop.Success {
				info, err := GetWhoisInfo(hop.Addr.String())
				if err == nil {
					printWhoisInfo(info)
				}
			}
			color.Normal.Println()
		}
	}
}

func printHop(hop Hop) {
	color.Normal.Printf("%2d ", hop.Number)
	if hop.Success {
		color.Yellow.Printf("%15s", hop.Addr.String())
		color.Normal.Printf(": ")
		color.Blue.Printf("%9s\n", hop.Rtt.Truncate(time.Microsecond))
	} else {
		color.Red.Printf("%15s\n", "*")
	}
}

func printErr(err error) {
	if strings.Contains(err.Error(), "operation not permitted") {
		color.LightRed.Println("Could not get network permissions")
		color.Normal.Println("Try running the program as root, or give it permission to use the network like this:")
		color.Normal.Println("$ sudo setcap cap_net_raw+ep tracert-as")
	} else {
		color.LightRed.Println(err)
	}
}

func printWhoisInfo(info WhoisInfo) {
	color.Normal.Print("   ")
	if info.Netname != "" {
		color.Green.Print(info.Netname)
	}
	if info.Netname != "" && info.AS != "" {
		color.Normal.Print(", ")
	}
	if info.AS != "" {
		color.Magenta.Print(info.AS)
	}
	if info.AS != "" || info.Netname != "" && info.Country != "" {
		color.Normal.Print(", ")
	}
	if info.Country != "" {
		color.Cyan.Print(info.Country)
	}
	color.Normal.Println()
}
