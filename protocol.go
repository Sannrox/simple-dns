package main

import (
	"fmt"
	"strings"
)

type BytePacketBuffer struct {
	buf []byte
	pos uint
}

func NewBytesPacketBuffer() *BytePacketBuffer {
	return &BytePacketBuffer{
		buf: make([]byte, 512),
		pos: 0,
	}
}

func (b *BytePacketBuffer) Pos() uint {
	return b.pos
}

func (b *BytePacketBuffer) Step(step uint) {
	b.pos += step
}

func (b *BytePacketBuffer) Seek(pos uint) {
	b.pos = pos
}

func (b *BytePacketBuffer) Read() (byte, error) {
	if b.pos >= 512 {
		return 0, fmt.Errorf("Read: end of buffer")
	}
	res := b.buf[b.pos]
	b.pos += 1

	return res, nil
}

func (b *BytePacketBuffer) Get(pos uint) (byte, error) {
	if pos >= 512 {
		return 0, fmt.Errorf("Get: end of buffer")
	}

	return b.buf[pos], nil
}

func (b *BytePacketBuffer) GetRange(start uint, len uint) ([]byte, error) {
	if start+len >= 512 {
		return nil, fmt.Errorf("Get range: end  of buffer")
	}

	return b.buf[start : start+len], nil
}

func (b *BytePacketBuffer) ReadU16() (uint16, error) {
	byte1, err := b.Read()
	if err != nil {
		return 0, fmt.Errorf("ReadU16: %s", err)
	}
	byte2, err := b.Read()
	if err != nil {
		return 0, fmt.Errorf("ReadU16: %s", err)
	}
	return (uint16(byte1) << 8) | uint16(byte2), nil
}

func (b *BytePacketBuffer) ReadU32() (uint32, error) {
	byte1, err := b.Read()
	if err != nil {
		return 0, fmt.Errorf("ReadU32: %s", err)
	}
	byte2, err := b.Read()
	if err != nil {
		return 0, fmt.Errorf("ReadU32: %s", err)
	}
	byte3, err := b.Read()
	if err != nil {
		return 0, fmt.Errorf("ReadU32: %s", err)
	}
	byte4, err := b.Read()
	if err != nil {
		return 0, fmt.Errorf("ReadU32: %s", err)
	}
	return (uint32(byte1) << 24) | (uint32(byte2) << 16) | (uint32(byte3) << 8) | uint32(byte4), nil
}

func (b *BytePacketBuffer) ReadQName(outString *string) error {

	var pos = b.Pos()

	var jumped bool = false
	var maxJumps int = 5
	var jumpsPerformed int = 0
	var builder strings.Builder
	var delimiter string = ""
	for {

		if jumpsPerformed > maxJumps {
			return fmt.Errorf("Limit of %d jumps exceeded", maxJumps)
		}

		len, err := b.Get(pos)
		if err != nil {
			return err
		}

		if (len & 0xC0) == 0xC0 {
			if !jumped {
				b.Seek(pos + 2)
			}

			b2, err := b.Get(pos + 1)
			if err != nil {
				return fmt.Errorf("ReadQName: %s", err)
			}
			var offset = ((uint16(len) ^ 0xC0) << 8) | uint16(b2)
			pos = uint(offset)

			jumped = true
			jumpsPerformed += 1

			continue

		} else {
			pos += 1

			if len == 0 {
				break
			}

			builder.WriteString(delimiter)

			strBuffer, err := b.GetRange(pos, uint(len))
			if err != nil {
				return fmt.Errorf("ReadQName: %s", err)

			}

			builder.Write(strBuffer)

			delimiter = "."

			pos += uint(len)

		}

	}
	if !jumped {
		b.Seek(pos)
	}
	*outString = builder.String()

	return nil

}

func (b *BytePacketBuffer) Write(val uint8) error {
	if b.pos >= 512 {
		return fmt.Errorf("Write: end of buffer")
	}
	b.buf[b.pos] = val
	b.pos += 1

	return nil
}

func (b *BytePacketBuffer) WriteU8(val uint8) error {
	return b.Write(val)
}

func (b *BytePacketBuffer) WriteU16(val uint16) error {
	err := b.Write(uint8(val >> 8))
	if err != nil {
		return err
	}
	return b.Write(uint8(val & 0xFF))
}

func (b *BytePacketBuffer) WriteU32(val uint32) error {
	if err := b.Write(uint8(val >> 24 & 0xFF)); err != nil {
		return err
	}
	if err := b.Write(uint8(val >> 16 & 0xFF)); err != nil {
		return err
	}
	if err := b.Write(uint8(val >> 8 & 0xFF)); err != nil {
		return err
	}
	return b.Write(uint8(val >> 0 & 0xFF))
}

func (b *BytePacketBuffer) WriteQName(qname *string) error {
	for _, label := range strings.Split(*qname, ".") {
		lenghtLable := len(label)
		if lenghtLable > 0x3F {
			return fmt.Errorf("Single label exceeds 63 characters")
		}

		err := b.Write(uint8(lenghtLable))
		if err != nil {
			return err
		}
		for _, c := range []byte(label) {
			err := b.Write(uint8(c))
			if err != nil {
				return err
			}
		}
	}
	return b.WriteU8(0)
}

func (b *BytePacketBuffer) Set(pos uint, val uint8) error {
	if pos >= 512 {
		return fmt.Errorf("Set: end of buffer")
	}
	b.buf[pos] = val
	return nil
}

func (b *BytePacketBuffer) SetU16(pos uint, val uint16) error {
	if err := b.Set(pos, uint8(val>>8)); err != nil {
		return err
	}
	return b.Set(pos+1, uint8(val&0xFF))
}
