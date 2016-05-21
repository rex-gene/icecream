package icinterface

import (
	"net"
)

type IClient interface {
	GetSrcSeq() uint16
	GetDstSeq() uint16
	GetToken() uint32
	GetAddr() *net.UDPAddr

	SetToken(uint32)
	SendData([]byte)
	IncDstSeq()

	SetSrcSeq(uint16)
}
