package socket

import (
	"github.com/RexGene/icecream/manager/datasendmanager"
	"github.com/RexGene/icecream/protocol"
	"net"
	"time"
)

type Socket struct {
	SrcSeq uint16
	DstSeq uint16

	Token uint32
	Addr  *net.UDPAddr

	state           int
	lastControlTime int64
	sender          *datasendmanager.DataSendManager
}

func New() *Socket {
	return &Socket{}
}

func (self *Socket) SetSender(sender *datasendmanager.DataSendManager) {
	self.sender = sender
}

func (self *Socket) Format(head *protocol.ICHead, addr *net.UDPAddr,
	sender *datasendmanager.DataSendManager) {
	self.Token = head.Token
	self.Addr = addr
	self.sender = sender
}

func (self *Socket) GetLastUpdateTime() int64 {
	return self.lastControlTime
}

func (self *Socket) GetState() int {
	return self.state
}

func (self *Socket) SetState(state int) {
	self.state = state
}

func (self *Socket) SendData(data []byte, size uint, isNeedBackup bool) {
	self.sender.SendData(self, data, size, isNeedBackup)
}

func (self *Socket) GetSrcSeq() uint16 {
	return self.SrcSeq
}

func (self *Socket) GetDstSeq() uint16 {
	return self.DstSeq
}

func (self *Socket) GetToken() uint32 {
	return self.Token
}

func (self *Socket) GetAddr() *net.UDPAddr {
	return self.Addr
}

func (self *Socket) SetToken(token uint32) {
	self.Token = token
}

func (self *Socket) IncDstSeq() {
	self.lastControlTime = time.Now().Unix()
	self.DstSeq++
}

func (self *Socket) IncSrcSeq() {
	self.lastControlTime = time.Now().Unix()
	self.SrcSeq++
}

func (self *Socket) SetSrcSeq(seq uint16) {
	self.lastControlTime = time.Now().Unix()
	self.SrcSeq = seq
}
