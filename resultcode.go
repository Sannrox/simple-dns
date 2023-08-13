package main

type ResultCode int

const (
	NOERROR ResultCode = iota // 0
	FORMERR
	SERVFAIL
	NXDOMAIN
	NOTIMP
	REFUSED
)
