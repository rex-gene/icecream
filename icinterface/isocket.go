package icinterface

import (
	"net"
)

const (
	SYN_STATE  = 0
	SYN_NORMAL = 1
	SHUT_DOWN  = 2
)

type ISocket interface {
	GetSrcSeq() uint16
	GetDstSeq() uint16
	GetToken() uint32
	GetAddr() *net.UDPAddr

	SetToken(uint32)
	SendData([]byte, uint, bool)

	IncDstSeq()
	IncSrcSeq()

	GetLastUpdateTime() int64

	GetState() int
	SetState(int)

	AddDstSeq(uint16)

	SetSrcSeq(uint16)
	InsertBackupList(uint16, []byte) uint
}
