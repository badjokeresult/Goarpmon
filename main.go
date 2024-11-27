package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
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

	parsed_table, _ := parseRawArpTableAsJson(results)

	data, err := json.Marshal(parsed_table)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	if _, e := os.Stdout.Write(data); e != nil {
		panic(e)
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

func parseRawArpTableAsJson(table []gosnmp.SnmpPDU) (*SnmpData, error) {
	results := []SnmpEntry{}
	for _, entry := range table {
		ipAddr, _ := formatIpAddr(entry.Name)
		value := hex.EncodeToString(entry.Value.([]byte))
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
