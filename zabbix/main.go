package zabbix

import (
	"encoding/json"
	"os"
)

func SendData(sender string, config string, key string, value interface{}) error {
	// data, err := json.Marshal(value)
	// if err != nil {
	// 	return err
	// }

	// cmd := exec.Command(sender, "-c", config, "-k", key, "-o", string(data[:]), "-vv >/var/log/arpmon/zabbix_sender.log 2>&1")

	// err = cmd.Run()
	// return err
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(data); err != nil {
		return err
	}
	return nil
}
