package main

import (
	"io/ioutil"
	"net"
	"strings"
	"time"
)

const (
	whoisPort = "43"
)

var (
	rirs = []string{
		"whois.ripe.net",
		"whois.arin.net",
		"whois.apnic.net",
		"whois.lacnic.net",
		"whois.afrinic.net",
	}
	badNetnames   = []string{"NON-RIPE-NCC-MANAGED-ADDRESS-BLOCK", "IANA-NETBLOCK", "ERX-NETBLOCK", "CIDR-BLOCK"}
	referField    = []string{"refer: ", "ReferralServer: "}
	netnameField  = []string{"NetName: ", "netname: "}
	originASField = []string{"OriginAS: ", "origin: ", "aut-num: "}
	countryField  = []string{"Country: ", "country: "}
)

// WhoisInfo contains information acquired from a whois request.
type WhoisInfo struct {
	Address, Netname, AS, Country string
}

// GetWhoisInfo returns information acquired from a whois request.
func GetWhoisInfo(host string) (info WhoisInfo, err error) {
	if isLocal(host) {
		return WhoisInfo{Netname: "local"}, nil
	}

	replies := make(chan string, len(rirs))
	for _, rir := range rirs {
		go getRawInfo(host, rir, replies)
	}

	maxFound := 0
	for range rirs {
		reply := <-replies
		result, found := parseInfo(host, reply)
		if found > maxFound {
			maxFound = found
			info = result
		}
	}
	return
}

func isLocal(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1]&0xf0 == 16) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	return len(ip) == 16 && ip[0]&0xfe == 0xfc
}

func parseInfo(host, reply string) (info WhoisInfo, found int) {
	var netname, as, country string

	if netname = parseField(reply, netnameField); netname != "" {
		found++
	}
	if as = parseField(reply, originASField); as != "" {
		found++
	}
	if country = parseField(reply, countryField); country != "" {
		found++
	}

	as = strings.TrimPrefix(as, "AS")

	if _, pos := findAny(netname, badNetnames); pos != -1 {
		return WhoisInfo{}, 0
	}

	return WhoisInfo{Address: host, Netname: netname, AS: as, Country: country}, found
}

func getRawInfo(host, server string, replies chan<- string) {
	reply, err := followReferral(host, server)
	if err != nil {
		replies <- ""
	}
	replies <- reply
}

func followReferral(host string, referral string) (reply string, err error) {
	reply, err = queryServer(host, referral)
	if err != nil {
		return "", err
	}
	foundServer := parseField(reply, referField)
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

	err = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	buffer, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

func findAny(str string, tokens []string) (token string, pos int) {
	for _, token := range tokens {
		pos = strings.Index(str, token)
		if pos != -1 {
			return token, pos
		}
	}
	return "", -1
}

func parseField(reply string, fieldVariants []string) (value string) {
	token, pos := findAny(reply, fieldVariants)
	if pos == -1 {
		return
	}
	pos += len(token) + 2
	len := strings.Index(reply[pos:], "\n")
	return strings.TrimSpace(reply[pos : pos+len])
}
