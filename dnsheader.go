package main

type DNSHeader struct {
	ID uint16

	recursionDesired bool
	truncatedMessage bool
	authoritative    bool
	opcode           uint8
	response         bool

	rescode            ResultCode
	checkingDisabled   bool
	authedData         bool
	z                  bool
	recursionAvailable bool

	questionCount        uint16
	answerCount          uint16
	authoritativeEntries uint16
	resourceEntries      uint16
}

func NewDNSHeader() *DNSHeader {
	return &DNSHeader{
		ID:               0,
		recursionDesired: false,
		truncatedMessage: false,
		authoritative:    false,
		opcode:           0,
		response:         false,

		rescode:            NOERROR,
		checkingDisabled:   false,
		authedData:         false,
		z:                  false,
		recursionAvailable: false,

		questionCount:        0,
		answerCount:          0,
		authoritativeEntries: 0,
		resourceEntries:      0,
	}
}

func (h *DNSHeader) Read(buffer *BytePacketBuffer) error {
	var err error
	h.ID, err = buffer.ReadU16()
	if err != nil {
		return err
	}

	flags, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	a := flags >> 8
	b := flags & 0x00FF

	h.recursionDesired = (a & (1 << 0)) > 0
	h.truncatedMessage = (a & (1 << 1)) > 0
	h.authoritative = (a & (1 << 2)) > 0
	h.opcode = uint8((a >> 3) & 0x0F)
	h.response = (a & (1 << 7)) > 0

	h.rescode = ResultCode(b & 0x0F)
	h.checkingDisabled = (b & (1 << 4)) > 0
	h.authedData = (b & (1 << 5)) > 0
	h.z = (b & (1 << 6)) > 0
	h.recursionAvailable = (b & (1 << 7)) > 0

	h.questionCount, err = buffer.ReadU16()
	if err != nil {
		return err
	}
	h.answerCount, err = buffer.ReadU16()
	if err != nil {
		return err
	}
	h.authoritativeEntries, err = buffer.ReadU16()
	if err != nil {
		return err
	}
	h.resourceEntries, err = buffer.ReadU16()
	if err != nil {
		return err
	}

	return nil

}

func (h *DNSHeader) Write(bufffer *BytePacketBuffer) error {
	if err := bufffer.WriteU16(h.ID); err != nil {
		return err
	}

	if err := bufffer.WriteU8(h.encodeFlagsA()); err != nil {
		return err
	}

	if err := bufffer.WriteU8(uint8(h.rescode)); err != nil {
		return err
	}

	if err := bufffer.WriteU16(h.questionCount); err != nil {
		return err
	}

	if err := bufffer.WriteU16(h.answerCount); err != nil {
		return err
	}

	if err := bufffer.WriteU16(h.authoritativeEntries); err != nil {
		return err
	}

	if err := bufffer.WriteU16(h.resourceEntries); err != nil {
		return err
	}

	return nil
}

func (h *DNSHeader) encodeFlagsA() uint8 {
	var flags uint8

	if h.recursionDesired {
		flags |= (1 << 0)
	}

	if h.truncatedMessage {
		flags |= (1 << 1)
	}

	if h.authoritative {
		flags |= (1 << 2)
	}

	flags |= (h.opcode << 3)

	if h.response {
		flags |= (1 << 7)
	}

	return flags
}

func (h *DNSHeader) encodeFlagsB() uint8 {
	var flags uint8

	if h.checkingDisabled {
		flags |= (1 << 4)
	}

	if h.authedData {
		flags |= (1 << 5)
	}

	if h.z {
		flags |= (1 << 6)
	}

	if h.recursionAvailable {
		flags |= (1 << 7)
	}

	return flags
}
