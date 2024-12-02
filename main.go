package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type SnmpData struct {
	Data []SnmpEntry `json:"data"`
}

type SnmpEntry struct {
	IpAddress string `json:"ipAddress"`
	HwAddress string `json:"hwAddress"`
	HostName  string `json:"hostName"`
}

func main() {
	host := os.Args[1]
	community := os.Args[2]
	port := uint16(161)

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

	parsed_table, err := parseRawArpTable(results, host)
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
