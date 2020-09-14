// Copyright (c) 2020 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// Released under the MIT license

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type IPScriptInput struct {
	IP string `json:"ip"`
}

type IPScriptOutput struct {
	Result     Action `json:"result"`
	BanMessage string `json:"banMessage"`
	// for caching: the network to which this result is applicable, and a TTL in seconds:
	CacheNet     string `json:"cacheNet"`
	CacheSeconds int    `json:"cacheSeconds"`
	Error        string `json:"error"`
}

func contains(i int, slice []int) bool {
	for _, j := range slice {
		if i == j {
			return true
		}
	}
	return false
}

type repliesConf struct {
	Codes  []int
	Action Action
	Reason string
}

type DNSBLConfigEntry struct {
	Host      string
	Addresses int
	Action    Action
	Reason    string
	Replies   []repliesConf
}

type Config struct {
	Precedence []Action
	Lists      []DNSBLConfigEntry
}

type Action int

const (
	IPAccepted    Action = 1
	IPBanned      Action = 2
	IPRequireSASL Action = 3
)

func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var orig string
	if err := unmarshal(&orig); err != nil {
		return err
	}
	switch strings.ToLower(orig) {
	case "allow", "accept":
		*a = IPAccepted
	case "block", "deny":
		*a = IPBanned
	case "require-sasl":
		*a = IPRequireSASL
	default:
		return fmt.Errorf("invalid action: %s", orig)
	}
	return nil
}

func evaluateDNSBL(conf DNSBLConfigEntry, ipv4 bool, reversedIP string, debug bool) (result Action, message string) {
	if (ipv4 && conf.Addresses == 6) || (!ipv4 && conf.Addresses == 4) {
		return IPAccepted, ""
	}

	hostname := reversedIP + conf.Host
	results, err := net.LookupHost(hostname)
	if err != nil || len(results) == 0 {
		if debug {
			fmt.Fprintf(os.Stderr, "%s returned no results\n", hostname)
		}
		return IPAccepted, ""
	}

	record := results[0]
	octets := strings.Split(record, ".")
	if debug {
		fmt.Fprintf(os.Stderr, "%s returned %s\n", hostname, record)
	}
	if len(octets) != 4 {
		if debug {
			fmt.Fprintf(os.Stderr, "corrupt response for %s: %s\n", hostname, record)
		}
		return IPAccepted, ""
	}
	code, err := strconv.Atoi(octets[3])
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "corrupt response for %s: %s\n", hostname, record)
		}
		return IPAccepted, ""
	}

	// see if this matches any of the special cased replies
	for i := range conf.Replies {
		if contains(code, conf.Replies[i].Codes) {
			return conf.Replies[i].Action, conf.Replies[i].Reason
		}
	}
	// ok, return the default
	return conf.Action, conf.Reason
}

func ReverseIP(ipaddr net.IP) (reversed string, ipv4 bool) {
	// include the trailing dot
	var b strings.Builder

	tofour := ipaddr.To4()
	if tofour != nil {
		ipv4 = true
		// 1.2.3.4 -> 4.3.2.1.dnsbl.domain
		for i := 3; i >= 0; i-- {
			fmt.Fprintf(&b, "%d.", tofour[i])
		}
	} else {
		for i := 15; i >= 0; i-- {
			octet := ipaddr[i]
			lsig_nibble := octet % 16
			msig_nibble := octet >> 4
			fmt.Fprintf(&b, "%s.", strconv.FormatInt(int64(lsig_nibble), 16))
			fmt.Fprintf(&b, "%s.", strconv.FormatInt(int64(msig_nibble), 16))
		}
	}

	return b.String(), ipv4
}

func LoadRawConfig(filename string) (config Config, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return
	}
	if len(config.Precedence) < 2 {
		config.Precedence = []Action{IPRequireSASL, IPBanned}
	}
	return
}

func run() (output IPScriptOutput, err error) {
	var ipaddr net.IP
	defer func() {
		output.BanMessage = strings.Replace(output.BanMessage, "{ip}", ipaddr.String(), -1)
	}()

	if len(os.Args) < 2 {
		err = fmt.Errorf("no config file supplied")
		return
	}
	debug := false
	if len(os.Args) > 2 {
		debug = true
	}
	config, err := LoadRawConfig(os.Args[1])
	if err != nil {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	var input IPScriptInput
	err = json.Unmarshal(line, &input)
	if err != nil {
		return
	}

	ipaddr = net.ParseIP(input.IP)
	if ipaddr == nil {
		err = fmt.Errorf("corrupt ip address %s", input.IP)
		return
	}
	reversed, ipv4 := ReverseIP(ipaddr)

	codes := make([]Action, len(config.Lists))
	reasons := make([]string, len(config.Lists))
	for i, list := range config.Lists {
		codes[i], reasons[i] = evaluateDNSBL(list, ipv4, reversed, debug)
		// fast path, if we got the highest precedence answer, no need to query any more
		if codes[i] == config.Precedence[0] {
			output.Result = codes[i]
			output.BanMessage = reasons[i]
			return
		}
	}

	for _, action := range config.Precedence {
		for i := 0; i < len(config.Lists); i++ {
			if codes[i] == action {
				output.Result = action
				output.BanMessage = reasons[i]
				return
			}
		}
	}

	output.Result = IPAccepted
	return
}

func main() {
	output, err := run()
	if err != nil {
		output.Result = 1 // allow
		output.Error = err.Error()
	}

	out, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}
	out = append(out, '\n')
	os.Stdout.Write(out)
}
