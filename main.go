package main

import (
	"fmt"
	"goarpmon/logger"
	"goarpmon/snmp"
	"goarpmon/zabbix"
	"log"
	"os"
	"strings"
)

func main() {
	logFileDir := os.Getenv("ARPMONDIR")
	logFileName := logFileDir + "/" + "arpmog.log"
	_, err := logger.StartLogging(logFileName, 512)
	if err != nil {
		log.Fatalf("Error start logging: %s", err)
	}

	host := os.Args[1]
	community := os.Args[2]
	arpDbFile := os.Getenv("ARPMONDB")

	arpTable := new(snmp.ArpTable)
	if err := arpTable.SetWithSnmpArpTableData(host, 161, community, arpDbFile); err != nil {
		log.Fatalf("Error retrieving ARP table: %s", err)
	}

	macIpsAddressesTable := new(snmp.AddressesTable)
	ipMacsAddressesTable := new(snmp.AddressesTable)

	macIpsAddressesTable.CalcIpAddressesDissonances(arpTable)
	ipMacsAddressesTable.CalcMacAddressesDissonances(arpTable)

	sender := "zabbix_sender"
	config := "/etc/zabbix/zabbix_sender.conf"
	logFilePath := logFileDir + "/" + "zabbix_sender.log"
	key := "arp.discovery"
	if err := zabbix.SendData(logFilePath, 512, sender, config, key, arpTable); err != nil {
		log.Fatalf("Error sending to Zabbix: %s", err)
	}
	for _, entry := range macIpsAddressesTable.Data {
		key = fmt.Sprintf("arp.macIps[%s]", entry.KeyAddr)
		value := strings.Join(entry.ValueAddrs, " ")
		if err := zabbix.SendData(logFilePath, 512, sender, config, key, value); err != nil {
			log.Fatalf("Error sending to Zabbix: %s", err)
		}
		key = fmt.Sprintf("arp.ipCount[%s]", entry.KeyAddr)
		val := len(entry.ValueAddrs)

	}

}
