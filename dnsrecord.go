package main

import (
	"fmt"
	"net"
)

type UnknownRecord struct {
	domain     string
	qtype      uint16
	dataLength uint16
	ttl        uint32
}

func (UnknownRecord) isDnsRecord() {}

func (UnknownRecord) Name() string {
	return "Unknown"
}

type ARecord struct {
	domain string
	addr   net.IP
	ttl    uint32
}

func (ARecord) isDnsRecord() {}

func (record ARecord) Name() string {
	return "A"
}

type NSRecord struct {
	domain string
	host   string
	ttl    uint32
}

func (NSRecord) isDnsRecord() {}

func (record NSRecord) Name() string {
	return "NS"
}

type CNameRecord struct {
	domain string
	host   string
	ttl    uint32
}

func (CNameRecord) isDnsRecord() {}

func (record CNameRecord) Name() string {
	return "CNAME"
}

type MXRecord struct {
	domain string
	prio   uint16
	host   string
	ttl    uint32
}

func (MXRecord) isDnsRecord() {}

func (record MXRecord) Name() string {
	return "MX"
}

type AAAARecord struct {
	domain string
	addr   net.IP
	ttl    uint32
}

func (AAAARecord) isDnsRecord() {}

func (record AAAARecord) Name() string {
	return "AAAA"
}

type DnsRecord interface {
	isDnsRecord()
	Name() string
}

func readDNSRecord(buffer *BytePacketBuffer) (DnsRecord, error) {
	var domain string
	err := buffer.ReadQName(&domain)
	if err != nil {
		return nil, fmt.Errorf("readDNSRecord.ReadQName: %s", err)
	}
	qtypeNum, err := buffer.ReadU16()
	if err != nil {
		return nil, fmt.Errorf("readDNSRecord.ReadU16.qtypeNum: %s", err)
	}
	qtype := QueryType(qtypeNum)
	_, err = buffer.ReadU16()
	if err != nil {
		return nil, fmt.Errorf("readDNSRecord.ReadU16.noname: %s", err)
	}
	ttl, err := buffer.ReadU32()
	if err != nil {
		return nil, fmt.Errorf("readDNSRecord.ReadU32.ttl: %s", err)
	}
	dataLength, err := buffer.ReadU16()
	if err != nil {
		return nil, fmt.Errorf("readDNSRecord.ReadU16.dataLength: %s", err)
	}
	switch qtype {
	case A:
		rawAddr, err := buffer.ReadU32()
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadU32.rawAddr: %s", err)
		}
		addr := net.IPv4(
			byte(rawAddr>>24),
			byte(rawAddr>>16),
			byte(rawAddr>>8),
			byte(rawAddr))
		return ARecord{domain, addr, ttl}, nil
	case AAAA:
		rawAddr1, err := buffer.ReadU32()
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadU16.rawAddr: %s", err)
		}
		rawAddr2, err := buffer.ReadU32()
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadU16.rawAddr: %s", err)
		}
		rawAddr3, err := buffer.ReadU32()
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadU16.rawAddr: %s", err)
		}
		rawAddr4, err := buffer.ReadU32()
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadU16.rawAddr: %s", err)
		}

		addr := net.IP{
			byte((rawAddr1 >> 24) & 0xFF), byte((rawAddr1 >> 16) & 0xFF), byte((rawAddr1 >> 8) & 0xFF), byte(rawAddr1 & 0xFF),
			byte((rawAddr2 >> 24) & 0xFF), byte((rawAddr2 >> 16) & 0xFF), byte((rawAddr2 >> 8) & 0xFF), byte(rawAddr2 & 0xFF),
			byte((rawAddr3 >> 24) & 0xFF), byte((rawAddr3 >> 16) & 0xFF), byte((rawAddr3 >> 8) & 0xFF), byte(rawAddr3 & 0xFF),
			byte((rawAddr4 >> 24) & 0xFF), byte((rawAddr4 >> 16) & 0xFF), byte((rawAddr4 >> 8) & 0xFF), byte(rawAddr4 & 0xFF),
		}
		return AAAARecord{domain, addr, ttl}, nil

	case NS:
		ns := ""
		err := buffer.ReadQName(&ns)
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadQName.ns: %s", err)
		}

		return NSRecord{domain, ns, ttl}, nil

	case CNAME:
		cname := ""
		err := buffer.ReadQName(&cname)
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadQName.cname: %s", err)
		}

		return CNameRecord{domain, cname, ttl}, nil
	case MX:
		prio, err := buffer.ReadU16()
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadU16.prio: %s", err)
		}

		mx := ""
		err = buffer.ReadQName(&mx)
		if err != nil {
			return nil, fmt.Errorf("readDNSRecord.ReadQName.mx: %s", err)
		}

		return MXRecord{domain, prio, mx, ttl}, nil

	default:
		buffer.Step(uint(dataLength))
		return UnknownRecord{domain, qtypeNum, dataLength, ttl}, nil

	}

}

func WriteDNSRecord(buffer *BytePacketBuffer, record DnsRecord) (uint, error) {
	startPos := buffer.Pos()
	switch record := record.(type) {
	case ARecord:
		if err := buffer.WriteQName(&record.domain); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		if err := buffer.WriteU16(uint16(A)); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.qtype: %s", err)
		}
		if err := buffer.WriteU16(1); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.class: %s", err)
		}
		if err := buffer.WriteU32(record.ttl); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU32.ttl: %s", err)
		}
		if err := buffer.WriteU16(4); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.dataLength: %s", err)
		}
		octets := []byte(record.addr)
		if err := buffer.WriteU8(octets[0]); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU8.octets[0]: %s", err)
		}
		if err := buffer.WriteU8(octets[1]); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU8.octets[1]: %s", err)
		}
		if err := buffer.WriteU8(octets[2]); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU8.octets[2]: %s", err)
		}

		if err := buffer.WriteU8(octets[3]); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU8.octets[3]: %s", err)
		}
	case NSRecord:
		if err := buffer.WriteQName(&record.domain); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		if err := buffer.WriteU16(uint16(NS)); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.qtype: %s", err)
		}

		if err := buffer.WriteU16(1); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.class: %s", err)
		}

		if err := buffer.WriteU32(record.ttl); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU32.ttl: %s", err)
		}

		pos := buffer.Pos()
		buffer.WriteU16(0)

		if err := buffer.WriteQName(&record.host); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		size := buffer.Pos() - (pos + 2)
		buffer.SetU16(pos, uint16(size))
	case CNameRecord:
		if err := buffer.WriteQName(&record.domain); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		if err := buffer.WriteU16(uint16(CNAME)); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.qtype: %s", err)
		}

		if err := buffer.WriteU16(1); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.class: %s", err)
		}

		if err := buffer.WriteU32(record.ttl); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU32.ttl: %s", err)
		}

		pos := buffer.Pos()
		buffer.WriteU16(0)

		if err := buffer.WriteQName(&record.host); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		size := buffer.Pos() - (pos + 2)
		buffer.SetU16(pos, uint16(size))
	case MXRecord:
		if err := buffer.WriteQName(&record.domain); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		if err := buffer.WriteU16(uint16(MX)); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.qtype: %s", err)
		}

		if err := buffer.WriteU16(1); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.class: %s", err)
		}

		if err := buffer.WriteU32(record.ttl); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU32.ttl: %s", err)
		}

		pos := buffer.Pos()
		buffer.WriteU16(0)

		if err := buffer.WriteU16(record.prio); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.prio: %s", err)
		}

		if err := buffer.WriteQName(&record.host); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		size := buffer.Pos() - (pos + 2)
		buffer.SetU16(pos, uint16(size))

	case AAAARecord:
		if err := buffer.WriteQName(&record.domain); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteQName: %s", err)
		}

		if err := buffer.WriteU16(uint16(AAAA)); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.qtype: %s", err)
		}

		if err := buffer.WriteU16(1); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.class: %s", err)
		}

		if err := buffer.WriteU32(record.ttl); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU32.ttl: %s", err)
		}

		if err := buffer.WriteU16(16); err != nil {
			return 0, fmt.Errorf("WriteDNSRecord.WriteU16.dataLength: %s", err)
		}

		octets := []byte(record.addr)
		for _, octet := range octets {
			if err := buffer.WriteU8(octet); err != nil {
				return 0, fmt.Errorf("WriteDNSRecord.WriteU8.octet: %s", err)
			}
		}

	default:
		fmt.Printf("Skipping record: %+v\n", record)
	}

	return buffer.Pos() - startPos, nil
}
