package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gosnmp/gosnmp"
)

type SnmpData struct {
	Data []SnmpEntry `json:"data"`
}

type SnmpEntry struct {
	IpAddress string `json:"ipAddress"`
	HwAddress string `json:"hwAddress"`
}

func main() {
	host := os.Args[1]
	community := os.Args[2]
	port := uint16(161)
	oid := "1.3.6.1.2.1.4.22.1.2"

	results, err := getRawArpTableBySNMP(host, port, community, oid)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	parsed_table, _ := parseRawArpTable(results)

	multuple_macs_on_single_ip, _ := calcMultipleMacsOnSingleIp(parsed_table)
	multiple_ips_on_signle_mac, _ := calcMultipleIpsOnSingleMac(parsed_table)

	config := "/etc/zabbix/zabbix_agentd.conf"
	sender := "/usr/bin/zabbix_sender"

	if err = sendDataToZabbix(sender, config, "arp.discovery", parsed_table); err != nil {
		log.Fatalf("Error: %s", err)
	}

	for key, value := range multuple_macs_on_single_ip {
		macsList := strings.Join(value, " ")
		if err = sendDataToZabbix(sender, config, fmt.Sprintf("arp.ipMacs[%s]", key), macsList); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}

	for key, value := range multiple_ips_on_signle_mac {
		ipsList := strings.Join(value, " ")
		if err = sendDataToZabbix(sender, config, fmt.Sprintf("arp.macIps[%s]", key), ipsList); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}
}

func getRawArpTableBySNMP(host string, port uint16, community string, oid string) ([]gosnmp.SnmpPDU, error) {
	g := &gosnmp.GoSNMP{
		Target:    host,
		Port:      port,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   gosnmp.Default.Timeout,
	}

	err := g.Connect()
	if err != nil {
		log.Fatalf("Error connecting")
	}
	defer g.Conn.Close()

	results, err := g.BulkWalkAll(oid)
	if err != nil {
		log.Fatalf("Error getting ARP")
	}

	return results, nil
}

func parseRawArpTable(table []gosnmp.SnmpPDU) (*SnmpData, error) {
	results := []SnmpEntry{}
	for _, entry := range table {
		ipAddr, _ := formatIpAddr(entry.Name)
		value := string(entry.Value.([]byte)[:])
		hwAddr, _ := formatMacAddr(string(value))
		results = append(results, SnmpEntry{
			IpAddress: ipAddr,
			HwAddress: hwAddr,
		})
	}

	return &SnmpData{
		Data: results,
	}, nil
}

func formatIpAddr(addr string) (string, error) {
	ipAddrArr := strings.Split(addr, ".")
	ipAddr := strings.Join(ipAddrArr[len(ipAddrArr)-4:], ".")
	return ipAddr, nil
}

func formatMacAddr(addr string) (string, error) {
	tmpArr := []string{}
	for i := 0; i < len(addr); i += 2 {
		tmpArr = append(tmpArr, addr[i:i+2])
	}
	mac := strings.Join(tmpArr, ":")
	return mac, nil
}

func calcMultipleMacsOnSingleIp(arp_table *SnmpData) (map[string][]string, error) {
	results := make(map[string][]string)
	for _, entry := range arp_table.Data {
		if _, ok := results[entry.HwAddress]; ok {
			results[entry.HwAddress] = append(results[entry.HwAddress], entry.IpAddress)
		} else {
			results[entry.HwAddress] = []string{}
		}
	}
	return results, nil
}

func calcMultipleIpsOnSingleMac(arp_table *SnmpData) (map[string][]string, error) {
	results := make(map[string][]string)
	for _, entry := range arp_table.Data {
		if _, ok := results[entry.IpAddress]; ok {
			results[entry.IpAddress] = append(results[entry.IpAddress], entry.HwAddress)
		} else {
			results[entry.IpAddress] = []string{}
		}
	}
	return results, nil
}

func sendDataToZabbix(sender string, config string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f, _ := os.OpenFile("./data.json", os.O_APPEND|os.O_CREATE, 0755)
	fmt.Fprintf(f, "%v\n\n\n", data[:])

	cmd := exec.Command(sender, "-c", config, "-k", key, "-o", string(data[:]), "-vv")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = os.Getenv("HOME")

	err = cmd.Run()
	return err
}
