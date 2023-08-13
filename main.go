package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	receivServer := "0.0.0.0:2053"
	localUDPAddr, err := net.ResolveUDPAddr("udp", receivServer)
	if err != nil {
		fmt.Println("Error resolving UDP address on ", receivServer)
		os.Exit(1)
	}
	receivConn, err := net.ListenUDP("udp", localUDPAddr)
	if err != nil {
		fmt.Println("Error listening on UDP port ", localUDPAddr)
		os.Exit(1)
	}

	defer receivConn.Close()

	fmt.Println("UDP server up and listening on port ", localUDPAddr.Port)
	for {
		if err := handleQuery(*receivConn); err != nil {
			fmt.Println("Error handling query", err)
		}

	}

}

func lookup(qname string, qtype QueryType) (*DNSPacket, error) {
	receivServer := "0.0.0.0:0"
	targetServer := "8.8.8.8:53"

	localUDPAddr, err := net.ResolveUDPAddr("udp", receivServer)
	if err != nil {
		fmt.Println("Error resolving UDP address on ", receivServer)
		os.Exit(1)
	}

	remoteUDPAddr, err := net.ResolveUDPAddr("udp", targetServer)
	if err != nil {
		fmt.Println("Error resolving UDP address on ", targetServer)
		os.Exit(1)
	}

	receivConn, err := net.ListenUDP("udp", localUDPAddr)
	if err != nil {
		fmt.Println("Error listening on UDP port ", localUDPAddr)
		os.Exit(1)
	}

	defer receivConn.Close()

	fmt.Println("UDP server up and listening on port ", localUDPAddr)

	packet := NewDNSPacket()
	packet.Header.ID = 6666
	packet.Header.questionCount = 1
	packet.Header.recursionDesired = true
	packet.Questions = append(packet.Questions, &DNSQuestion{qname, qtype})

	buffer := NewBytesPacketBuffer()
	if err := packet.Write(buffer); err != nil {
		fmt.Println("Error writing to buffer", err)
		os.Exit(1)
	}

	if _, err := receivConn.WriteToUDP(buffer.buf[:buffer.pos], remoteUDPAddr); err != nil {
		fmt.Println("Error writing to socket", err)
		os.Exit(1)
	}

	fmt.Printf("Sent packet with ID %d to %s\n", packet.Header.ID, targetServer)

	receivBuffer := NewBytesPacketBuffer()
	_, _, err = receivConn.ReadFromUDP(receivBuffer.buf)
	if err != nil {
		fmt.Println("Error reading from socket", err)
		os.Exit(1)
	}

	receivPacket := NewDNSPacket()
	receivPacket, err = receivPacket.Read(receivBuffer)
	if err != nil {
		fmt.Println("Error reading from buffer", err)
		os.Exit(1)
	}

	return receivPacket, nil
}

func handleQuery(socketConn net.UDPConn) error {
	reqBuffer := NewBytesPacketBuffer()
	_, src, err := socketConn.ReadFromUDP(reqBuffer.buf)
	if err != nil {
		return fmt.Errorf("Error reading from socket", err)
	}

	fmt.Printf("Received %d bytes from %s\n", reqBuffer.pos, src)

	reqPacket := NewDNSPacket()
	reqPacket, err = reqPacket.Read(reqBuffer)
	if err != nil {
		return fmt.Errorf("Error reading from buffer", err)
	}

	fmt.Printf("Received packet with ID %d\n", reqPacket.Header.ID)

	respPacket := NewDNSPacket()
	respPacket.Header.ID = reqPacket.Header.ID
	respPacket.Header.recursionDesired = true
	respPacket.Header.recursionAvailable = true
	respPacket.Header.response = true

	if len(reqPacket.Questions) > 0 {
		for _, q := range reqPacket.Questions {
			if packet, err := lookup(q.Name, q.Type); err != nil {
				respPacket.Header.rescode = SERVFAIL
			} else {
				respPacket.Questions = append(respPacket.Questions, q)
				respPacket.Header.rescode = packet.Header.rescode
				respPacket.Answers = append(respPacket.Answers, packet.Answers...)
				respPacket.Authorities = append(respPacket.Authorities, packet.Authorities...)
				respPacket.Reources = append(respPacket.Reources, packet.Reources...)
			}
		}
	} else {
		respPacket.Header.rescode = FORMERR
	}

	respBuffer := NewBytesPacketBuffer()
	if err := respPacket.Write(respBuffer); err != nil {
		return fmt.Errorf("Error writing to buffer", err)
	}

	len := respBuffer.Pos()
	data, err := respBuffer.GetRange(0, len)
	if err != nil {
		return fmt.Errorf("Error getting range", err)
	}

	if _, err := socketConn.WriteToUDP(data, src); err != nil {
		return fmt.Errorf("Error writing to socket", err)
	}

	return nil

}
