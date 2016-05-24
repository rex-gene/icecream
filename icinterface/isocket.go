package icinterface

import (
	"net"
)

const (
	SYN_STATE  = 0
	SYN_NORMAL = 1
)

type ISocket interface {
	GetSrcSeq() uint16
	GetDstSeq() uint16
	GetToken() uint32
	GetAddr() *net.UDPAddr

	SetToken(uint32)
	SendData([]byte, bool)

	IncDstSeq()
	IncSrcSeq()

	GetLastUpdateTime() int64

	GetState() int
	SetState(int)

	SetSrcSeq(uint16)
}
