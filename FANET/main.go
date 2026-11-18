package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

// The first byte of every packet says what kind of message it is.
const (
	msgJoinReq    byte = 1
	msgJoinAccept byte = 2
	msgJoinReject byte = 3
)

func main() {
	mode := flag.String("mode", "ch", "ch | drone")
	listen := flag.String("listen", ":9000", "address the CH listens on")
	chAddr := flag.String("ch", "127.0.0.1:9000", "address of the CH (used by drone)")
	flag.Parse()

	switch *mode {
	case "ch":
		runClusterHead(*listen)
	case "drone":
		runDrone(*chAddr)
	default:
		log.Fatalf("unknown mode %q", *mode)
	}
}

func runClusterHead(listen string) {
	addr, err := net.ResolveUDPAddr("udp", listen)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("[CH] listening on", listen)

	buf := make([]byte, 65535) // max UDP size; PQC packets will be big later
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		if n < 1 || buf[0] != msgJoinReq {
			fmt.Println("[CH] ignored unknown packet from", remote)
			continue
		}
		fmt.Println("[CH] JOIN_REQ from", remote)

		conn.WriteToUDP([]byte{msgJoinAccept}, remote)
		fmt.Println("[CH] sent JOIN_ACCEPT to", remote)
	}
}

func runDrone(chAddr string) {
	remote, err := net.ResolveUDPAddr("udp", chAddr)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp", nil, remote)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conn.Write([]byte{msgJoinReq})
	fmt.Println("[DRONE] sent JOIN_REQ to", chAddr, "- waiting...")

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("[DRONE] no answer - flying alone")
		return
	}

	if n >= 1 && buf[0] == msgJoinAccept {
		fmt.Println("[DRONE] accepted into the cluster!")
	} else {
		fmt.Println("[DRONE] rejected or unknown reply")
	}
}
