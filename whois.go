package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

const (
	ianaWhoisServer = "whois.iana.org"
	whoisPort       = "43"
)

var (
	referField    = []string{"refer", "ReferralServer"}
	netnameField  = []string{"NetName", "netname"}
	originASField = []string{"OriginAS", "origin", "aut-num"}
	countryField  = []string{"Country", "country"}
)

// WhoisInfo contains information acquired from a whois request.
type WhoisInfo struct {
	Address, Netname, AS, Country string
}

// GetWhoisInfo returns information acquired from a whois request.
func GetWhoisInfo(host string) (WhoisInfo, error) {
	reply, err := getRawInfo(host)
	if err != nil {
		return WhoisInfo{}, err
	}

	netname := parseField(reply, netnameField)
	as := parseField(reply, originASField)
	country := parseField(reply, countryField)

	as = strings.TrimPrefix(as, "AS")

	return WhoisInfo{Address: host, Netname: netname, AS: as, Country: country}, nil
}

func getRawInfo(host string) (reply string, err error) {
	server, reply, err := followReferral(host, ianaWhoisServer)
	if err != nil {
		return "", err
	}
	if server == ianaWhoisServer {
		return "", fmt.Errorf("Could not find the whois server")
	}
	return
}

func followReferral(host string, referral string) (foundServer, reply string, err error) {
	reply, err = queryServer(host, referral)
	if err != nil {
		return "", "", err
	}
	foundServer = parseField(reply, referField)
	if i := strings.LastIndex(foundServer, "/"); i != -1 {
		foundServer = foundServer[i+1:]
	}
	if strings.Index(foundServer, "rwhois") != -1 {
		foundServer = ""
	}
	if foundServer == "" {
		return
	}
	return followReferral(host, foundServer)
}

func queryServer(host string, server string) (reply string, err error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(server, whoisPort))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(host + "\r\n"))
	if err != nil {
		return "", err
	}

	err = conn.SetReadDeadline(time.Now().Add(time.Second / 2))
	buffer, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

func parseField(reply string, fieldVariants []string) (value string) {
	for _, field := range fieldVariants {
		token := field + ": "
		pos := strings.Index(reply, token)
		if pos != -1 {
			pos += len(token)
			len := strings.Index(reply[pos:], "\n")
			return strings.TrimSpace(reply[pos : pos+len])
		}
	}
	return ""
}
