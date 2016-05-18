package client

import (
	"github.com/RexGene/icecream/protocol"
	"net"
)

type Client struct {
	scrSeq uint16
	dstSeq uint16

	token uint32
	addr  *net.UDPAddr
}

func New() *Client {

}

func (self *Client) Format(head *protocol.ICHead, addr *net.UDPAddr) {
	self.token = head.Token
	self.addr = addr
	self.scrSeq = 0
	self.dstSeq = 0
}

func (self *Client) SendData(data []byte) {
}
