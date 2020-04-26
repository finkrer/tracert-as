package main

import (
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
	err := TraceRoute(host)
	if err != nil {
		color.Red.Println(err)
	}
}
