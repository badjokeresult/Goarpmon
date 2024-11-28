package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
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

	logf, err := os.OpenFile("arpmon.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer logf.Close()
	log.SetOutput(logf)

	results, err := getRawArpTableBySNMP(host, port, community, oid)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Printf("ARP-table was received from %s by SNMPv2c", host)

	parsed_table, err := parseRawArpTable(results)
	if err != nil {
		log.Fatalf("Error parsing raw ARP-table: %s", err)
	}
	log.Printf("ARP-table was successfully parsed")

	multiple_ips_on_single_mac, multiple_macs_on_single_ip, err := calcExtraAddresses(parsed_table)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Printf("Addresses' dissonances were calculated successfully")

	config := "/etc/zabbix/zabbix_agentd.conf"
	sender := "/usr/bin/zabbix_sender"
	if err = sendDataToZabbix(sender, config, "arp.discovery", parsed_table); err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Printf("ARP-table was sent to Zabbix successfully")

	for key, value := range multiple_macs_on_single_ip {
		macsList := strings.Join(value, " ")
		macCount := strconv.Itoa(len(value))
		if err = sendDataToZabbix(sender, config, fmt.Sprintf("arp.ipMacs[%s]", key), macsList); err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = sendDataToZabbix(sender, config, fmt.Sprintf("arp.macCount[%s]", key), macCount); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}
	log.Printf("\"Multiple MACs Single IP\" dissonances were sent to Zabbix successfully")

	for key, value := range multiple_ips_on_single_mac {
		ipsList := strings.Join(value, " ")
		ipCount := strconv.Itoa(len(value))

		if err = sendDataToZabbix(sender, config, fmt.Sprintf("arp.macIps[%s]", key), ipsList); err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = sendDataToZabbix(sender, config, fmt.Sprintf("arp.ipCount[%s]", key), ipCount); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}
	log.Printf("\"Multiple IPs Single MAC\" dissonances were sent to Zabbix successfully")

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
		return nil, err
	}
	defer g.Conn.Close()

	results, err := g.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func parseRawArpTable(table []gosnmp.SnmpPDU) (*SnmpData, error) {
	results := []SnmpEntry{}
	for _, entry := range table {
		ipAddr, _ := formatIpAddr(entry.Name)
		if hwAsByteArr, ok := entry.Value.([]byte); ok {
			hwAddr, _ := formatMacAddr(hwAsByteArr)
			results = append(results, SnmpEntry{
				IpAddress: ipAddr,
				HwAddress: hwAddr,
			})
		}
	}

	return &SnmpData{
		Data: results[:],
	}, nil
}

func formatIpAddr(addr string) (string, error) {
	ipAddrArr := strings.Split(addr, ".")
	ipAddr := strings.Join(ipAddrArr[len(ipAddrArr)-4:], ".")
	return ipAddr, nil
}

func formatMacAddr(addr []byte) (string, error) {
	result := []string{}
	for i := 0; i < len(addr); i++ {
		result = append(result, fmt.Sprintf("%02X", addr[i]))
	}
	return strings.Join(result, ":"), nil
}

func calcExtraAddresses(arpTable *SnmpData) (map[string][]string, map[string][]string, error) {
	macIps := make(map[string][]string)
	for _, entry := range arpTable.Data {
		if _, ok := macIps[entry.HwAddress]; !ok {
			macIps[entry.HwAddress] = []string{}
			for _, inner := range arpTable.Data {
				if entry.HwAddress == inner.HwAddress {
					macIps[entry.HwAddress] = append(macIps[entry.HwAddress], inner.IpAddress)
				}
			}
		}
	}

	ipMacs := make(map[string][]string)
	for _, entry := range arpTable.Data {
		if _, ok := ipMacs[entry.IpAddress]; !ok {
			ipMacs[entry.IpAddress] = []string{}
			for _, inner := range arpTable.Data {
				if entry.IpAddress == inner.IpAddress {
					ipMacs[entry.IpAddress] = append(ipMacs[entry.IpAddress], entry.HwAddress)
				}
			}
		}
	}

	fmt.Fprintln(os.Stdout, macIps, ipMacs)
	return macIps, ipMacs, nil
}

func sendDataToZabbix(sender string, config string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	cmd := exec.Command(sender, "-c", config, "-k", key, "-o", string(data[:]), "-vv")

	f, err := os.OpenFile("zabbix_sender.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Dir = os.Getenv("HOME")

	err = cmd.Run()
	return err
}
