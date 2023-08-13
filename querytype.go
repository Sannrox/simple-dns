package main

type QueryType uint16

const (
	UNKNOWN QueryType = iota
	A       QueryType = 1
	NS      QueryType = 2
	CNAME   QueryType = 5
	MX      QueryType = 15
	AAAA    QueryType = 28
)

func (qt QueryType) String() string {
	switch qt {
	case A:
		return "A"
	case NS:
		return "NS"
	case CNAME:
		return "CNAME"
	case MX:
		return "MX"
	case AAAA:
		return "AAAA"
	default:
		return "UNKNOWN"
	}
}
