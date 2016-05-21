package protocol

const (
	START_FLAG = 1 << 4
	STOP_FLAG  = 1 << 3
	RESET_FLAG = 1 << 2
	PUSH_FLAG  = 1 << 1
	ACK_FLAG   = 1 << 0
)

type ICHead struct {
	Flag     byte
	Sum      byte
	SrcSeqId uint16
	DstSeqId uint16
	Len      uint16
	Token    uint32
	CmdId    uint32
}
