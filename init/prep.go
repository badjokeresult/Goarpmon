package init

import (
	"os"
)

func CreateConfDir(path string) error {
	err := os.MkdirAll(path, 0644)
	if err != nil {
		return err
	}
}

func LaunchArpScan()

func InitializeArpMon()
