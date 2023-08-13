package main

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
)

type DNSPacket struct {
	Header      *DNSHeader
	Questions   []*DNSQuestion
	Answers     []DnsRecord
	Authorities []DnsRecord
	Reources    []DnsRecord
}

func NewDNSPacket() *DNSPacket {
	return &DNSPacket{
		Header:      NewDNSHeader(),
		Questions:   make([]*DNSQuestion, 0),
		Answers:     make([]DnsRecord, 0),
		Authorities: make([]DnsRecord, 0),
		Reources:    make([]DnsRecord, 0),
	}
}

func (d *DNSPacket) Read(buffer *BytePacketBuffer) (*DNSPacket, error) {
	result := NewDNSPacket()
	err := result.Header.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("DNSPacket.Read.Header: %s", err)
	}

	for i := 0; i < int(result.Header.questionCount); i++ {
		question := NewDNSQuestion("", UNKNOWN)
		err := question.Read(buffer)
		if err != nil {
			return nil, fmt.Errorf("DNSPacket.Read.Questions: %s", err)
		}
		result.Questions = append(result.Questions, question)
	}

	for i := 0; i < int(result.Header.answerCount); i++ {
		record, err := readDNSRecord(buffer)
		if err != nil {
			return nil, fmt.Errorf("DNSPacket.Read.Answers: %s", err)
		}
		result.Answers = append(result.Answers, record)
	}

	for i := 0; i < int(result.Header.authoritativeEntries); i++ {
		record, err := readDNSRecord(buffer)
		if err != nil {
			return nil, fmt.Errorf("DNSPacket.Read.Authorities: %s", err)
		}
		result.Authorities = append(result.Authorities, record)
	}

	for i := 0; i < int(result.Header.resourceEntries); i++ {
		record, err := readDNSRecord(buffer)
		if err != nil {
			return nil, fmt.Errorf("DNSPacket.Read.Reources: %s", err)
		}
		result.Reources = append(result.Reources, record)
	}

	return result, nil
}

func (d *DNSPacket) Write(buffer *BytePacketBuffer) error {
	d.Header.questionCount = uint16(len(d.Questions))
	d.Header.answerCount = uint16(len(d.Answers))
	d.Header.authoritativeEntries = uint16(len(d.Authorities))
	d.Header.resourceEntries = uint16(len(d.Reources))

	if err := d.Header.Write(buffer); err != nil {
		return fmt.Errorf("DNSPacket.Write.Header: %s", err)
	}

	for _, question := range d.Questions {
		if err := question.Write(buffer); err != nil {
			return fmt.Errorf("DNSPacket.Write.Questions: %s", err)
		}
	}

	for _, record := range d.Answers {
		if _, err := WriteDNSRecord(buffer, record); err != nil {
			return fmt.Errorf("DNSPacket.Write.Answers: %s", err)
		}
	}

	for _, record := range d.Authorities {
		if _, err := WriteDNSRecord(buffer, record); err != nil {
			return fmt.Errorf("DNSPacket.Write.Authorities: %s", err)
		}

	}

	for _, record := range d.Reources {
		if _, err := WriteDNSRecord(buffer, record); err != nil {
			return fmt.Errorf("DNSPacket.Write.Reources: %s", err)
		}
	}

	return nil
}

func (d *DNSPacket) GetRandomA() (net.IP, error) {

	var answers []net.IP

	for _, answer := range d.Answers {
		if record, ok := answer.(*ARecord); ok {
			answers = append(answers, record.addr)
		}
	}
	if len(answers) > 0 {
		randomIndex := rand.Intn(len(answers))
		return answers[randomIndex], nil
	}

	return nil, fmt.Errorf("DNSPacket.GetRandomA: No A records found")
}

func (d *DNSPacket) GetNS(qname string) <-chan struct{ NSDomain, NSHost string } {
	output := make(chan struct{ NSDomain, NSHost string })

	go func() {
		defer close(output)

		for _, record := range d.Authorities {
			if record, ok := record.(*NSRecord); ok {

				if record.domain != "" && record.host != "" {
					if strings.HasSuffix(qname, record.domain) {
						output <- struct{ NSDomain, NSHost string }{NSDomain: record.domain, NSHost: record.host}
					}
				}
			}
		}
	}()

	return output
}

func (d *DNSPacket) GetResolvedNS(qname string) net.IP {
	var resolvedIP net.IP
	nsRecords := d.GetNS(qname)

	for nsRecord := range nsRecords {
		host := nsRecord.NSHost

		// Look for a matching A record in the additional section
		for _, record := range d.Reources {
			if record, ok := record.(*ARecord); ok {

				if record.domain == host && record.addr != nil {
					resolvedIP = record.addr
					break
				}
			}
		}

		if resolvedIP != nil {
			break
		}
	}

	return resolvedIP

}
