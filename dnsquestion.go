package main

import "fmt"

type DNSQuestion struct {
	Name string
	Type QueryType
}

func NewDNSQuestion(name string, qt QueryType) *DNSQuestion {
	return &DNSQuestion{
		Name: name,
		Type: qt,
	}
}

func (q *DNSQuestion) Read(b *BytePacketBuffer) error {
	err := b.ReadQName(&q.Name)
	if err != nil {
		return err
	}

	by, err := b.ReadU16()
	if err != nil {
		return err
	}

	q.Type = QueryType(by)
	_, err = b.ReadU16()
	if err != nil {
		return err
	}

	return nil
}

func (q *DNSQuestion) Write(b *BytePacketBuffer) error {
	if err := b.WriteQName(&q.Name); err != nil {
		return err
	}

	if err := b.WriteU16(uint16(q.Type)); err != nil {
		return err
	}

	if err := b.WriteU16(1); err != nil {
		return err
	}

	return nil
}

func (q *DNSQuestion) String() string {
	return fmt.Sprintf("%s %s", q.Name, q.Type)
}
