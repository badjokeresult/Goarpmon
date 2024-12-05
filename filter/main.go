package filter

import (
	"os"
	"strings"
)

type LegitMacAddressesTable struct {
	Table []string
}

func (t *LegitMacAddressesTable) RetrieveAddresses(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer f.Close()

	buf := []byte{}
	if _, err := f.Read(buf); err != nil {
		return err
	}

	table := string(buf[:])
	lines := strings.Split(table, "\n")
	for _, line := range lines {
		addrPair := strings.Split(line, "\t")
		t.Table = append(t.Table, addrPair[1])
	}

	return nil
}

func (t *LegitMacAddressesTable) IsAddrInTable(addr string) bool {
	for _, a := range t.Table {
		if addr == a {
			return true
		}
	}
	return false
}
