package main

import (
	"flag"
	"fmt"
	"goarpmon/config"
	"goarpmon/snmp"
	"goarpmon/zabbix"
	"log"
	"os"
	"strings"
)

func main() {
	conf, err := config.Parse("/etc/arpmon/config.toml")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Println("Config file `/etc/arpmon/config.toml` was parsed successfully")

	host := flag.String("host", "0.0.0.0", "Network device IPv4-address or hostname")
	community := flag.String("community", "community", "SNMPv2c community name")
	flag.Parse()

	logFileName := conf.Arp.Log
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("Error start logging: %s", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("Logging was started successfully")

	arpDbFile := conf.Arp.Db
	arpTable := new(snmp.ArpTable)
	if err := arpTable.SetWithSnmpArpTableData(*host, 161, *community, arpDbFile); err != nil {
		log.Fatalf("Error retrieving ARP table: %s", err)
	}
	log.Println("ARP table was retrieved successfully")

	macIpsAddressesTable := new(snmp.AddressesTable)
	ipMacsAddressesTable := new(snmp.AddressesTable)

	macIpsAddressesTable.CalcIpAddressesDissonances(arpTable)
	ipMacsAddressesTable.CalcMacAddressesDissonances(arpTable)

	sender := conf.Zabbix.Sender
	config := conf.Zabbix.Config
	zabbixLog := conf.Zabbix.Log
	key := "arp.discovery"
	if err := zabbix.SendData(sender, config, key, zabbixLog, arpTable); err != nil {
		log.Fatalf("Error sending to Zabbix: %s", err)
	}
	log.Printf("`arp.discovery` was sent successfully. Data: %v", arpTable)
	for _, entry := range macIpsAddressesTable.Data {
		key = fmt.Sprintf("arp.macIps[%s]", entry.KeyAddr)
		value := strings.Join(entry.ValueAddrs, " ")
		if err := zabbix.SendData(sender, config, key, zabbixLog, value); err != nil {
			log.Fatalf("Error sending to Zabbix: %s", err)
		}
		log.Printf("`%s` was sent successfully. Data: %v", key, value)
		key = fmt.Sprintf("arp.ipCount[%s]", entry.KeyAddr)
		val := len(entry.ValueAddrs)
		if err := zabbix.SendData(sender, config, key, zabbixLog, val); err != nil {
			log.Fatalf("Error sending to Zabbix: %s", err)
		}
		log.Printf("`%s` was sent successfully. Data: %v", key, val)
	}
	for _, entry := range ipMacsAddressesTable.Data {
		key = fmt.Sprintf("arp.ipMacs[%s]", entry.KeyAddr)
		value := strings.Join(entry.ValueAddrs, " ")
		if err := zabbix.SendData(sender, config, key, zabbixLog, value); err != nil {
			log.Fatalf("Error sending to Zabbix: %s", err)
		}
		log.Printf("`%s` was sent successfully. Data: %v", key, value)
		key = fmt.Sprintf("arp.macCount[%s]", entry.KeyAddr)
		val := len(entry.ValueAddrs)
		if err := zabbix.SendData(sender, config, key, zabbixLog, val); err != nil {
			log.Fatalf("Error sending to Zabbix: %s", err)
		}
		log.Printf("`%s` was sent successfully. Data: %v", key, val)
	}
}
