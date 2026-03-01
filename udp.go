package avcamx

import (
	"log"
	"net"
	"time"
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

func PollUDP(done chan int, updateAddr chan string) error {

	var err error
	localAddr := GetOutboundIP()
	udpAddr, err := net.ResolveUDPAddr("udp4", UDPAddress())
	if err != nil {
		log.Println("ResolveUDPAddr: ", err)
		return err
	}
	// Start listening for UDP packages on the given address
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Println("ListenUDP: ", err)
		return err
	}

	var (
		buf  [1024]byte
		addr *net.UDPAddr
	)

	for {
		select {
		case <-done:
			return nil
		default:
			time.Sleep(time.Second)
		}

		_, addr, err = conn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Println("ReadFromUDP: ", err)
			continue
		}

		remoteAddr := addr.IP.String()
		if remoteAddr == localAddr {
			continue
		}

		log.Print("> ", string(buf[0:]), remoteAddr)
		updateAddr <- remoteAddr
		// // Write back the message over UPD
		// conn.WriteToUDP([]byte("Hello UDP Client\n"), addr)
	}

	// return nil
}
