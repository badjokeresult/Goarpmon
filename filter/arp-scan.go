package filter

import (
	"bufio"
	"os"
	"strings"
)

type LegitMacAddressesTable struct {
	Table []string
}

func (t *LegitMacAddressesTable) RetrieveAddresses(fileName string) error {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0640)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		mac := strings.Split(scanner.Text(), "\t")[1]
		t.Table = append(t.Table, mac)
	}

	return scanner.Err()
}

func (t *LegitMacAddressesTable) IsAddrInTable(addr string) bool {
	for _, a := range t.Table {
		if addr == a {
			return true
		}
	}
	return false
}
