package avcamx

import (
	"log"
	"net"
)

const UDPPort = ":9010"

func UDPAddress() string {
	local := GetOutboundIP()
	return local[:len(local)-4] + ".255" + UDPPort
}

// thanks to Mr. Wong
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func DialUDP(msg string) (err error) {
	var conn net.Conn
	conn, err = net.Dial("udp4", UDPAddress())
	if err != nil {
		log.Printf("DialUDP %v", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(msg))
	return
}
