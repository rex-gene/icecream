package client

import (
	"github.com/RexGene/common/objectpool"
	"github.com/RexGene/icecream/manager/datasendmanager"
	"github.com/RexGene/icecream/protocol"
	"net"
)

type Client struct {
	objectpool.BaseObject
	SrcSeq uint16
	DstSeq uint16

	Token uint32
	Addr  *net.UDPAddr
}

func New() *Client {
	return &Client{}
}

func (self *Client) Format(head *protocol.ICHead, addr *net.UDPAddr) {
	self.Token = head.Token
	self.Addr = addr
	self.SrcSeq = 0
	self.DstSeq = 0
}

func (self *Client) SendData(data []byte) {
	self.SrcSeq++
	datasendmanager.GetInstance().SendData(self, data)
}

func (self *Client) GetSrcSeq() uint16 {
	return self.SrcSeq
}

func (self *Client) GetDstSeq() uint16 {
	return self.DstSeq
}

func (self *Client) GetToken() uint32 {
	return self.Token
}

func (self *Client) GetAddr() *net.UDPAddr {
	return self.Addr
}

func (self *Client) SetToken(token uint32) {
	self.Token = token
}

func (self *Client) IncDstSeq() {
	self.DstSeq++
}

func (self *Client) SetSrcSeq(seq uint16) {
	self.SrcSeq = seq
}
