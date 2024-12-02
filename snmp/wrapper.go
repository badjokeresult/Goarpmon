package snmp

import (
	"fmt"
	"strings"

	"github.com/gosnmp/gosnmp"
)

const OID string = "1.3.6.1.2.1.4.22.1.2"

type ArpTable struct {
	Data []arpEntry `json:"data"`
}

type arpEntry struct {
	IpAddress string `json:"ipAddress"`
	HwAddress string `json:"hwAddress"`
	HostName  string `json:"hostName"`
}

type AddressesTable struct {
	Data []addressesEntry
}

type addressesEntry struct {
	KeyAddr    string
	ValueAddrs []string
}

func (a *ArpTable) SetWithSnmpArpTableData(host string, port uint16, community string) error {
	rawArpTable, err := getRawArpTableBySNMP(host, port, community, OID)
	if err != nil {
		return err
	}

	a.fillArpTableFields(rawArpTable, host)
	return nil
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

func (a *ArpTable) fillArpTableFields(table []gosnmp.SnmpPDU, hostName string) {
	results := []arpEntry{}
	for _, entry := range table {
		ipAddr := formatIpAddr(entry.Name)
		if hwAsByteArr, ok := entry.Value.([]byte); ok {
			hwAddr := formatMacAddr(hwAsByteArr)
			results = append(results, arpEntry{
				IpAddress: ipAddr,
				HwAddress: hwAddr,
				HostName:  hostName,
			})
		}
	}

	a.Data = results
}

func formatIpAddr(addr string) string {
	ipAddrArr := strings.Split(addr, ".")
	ipAddr := strings.Join(ipAddrArr[len(ipAddrArr)-4:], ".")
	return ipAddr
}

func formatMacAddr(addr []byte) string {
	result := []string{}
	for i := 0; i < len(addr); i++ {
		result = append(result, fmt.Sprintf("%02X", addr[i]))
	}
	return strings.Join(result, ":")
}

func CalcIpAddressesDissonances(arpTable *ArpTable) (map[string][]string, map[string][]string) {
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

	return macIps, ipMacs
}

func (a *AddressesTable) CalcMacAddressesDissonances(arpTable *ArpTable) {
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

	for k, v := range ipMacs {
		a.Data = append(a.Data, addressesEntry{
			KeyAddr:    k,
			ValueAddrs: v,
		})
	}
}
