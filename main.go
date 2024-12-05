package main

type SnmpData struct {
	Data []SnmpEntry `json:"data"`
}

type SnmpEntry struct {
	IpAddress string `json:"ipAddress"`
	HwAddress string `json:"hwAddress"`
	HostName  string `json:"hostName"`
}

func main() {

}
