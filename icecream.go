package icecream

import (
	"github.com/RexGene/common/memorypool"
	_ "github.com/RexGene/common/objectpool"
	"github.com/RexGene/common/threadpool"
	"log"
	"net"
)

const (
	MAX_CONNECT_COUNT = 1024
	READ_BUFFER_SIZE  = 65535
)

type IceCream struct {
	udpAddr   *net.UDPAddr
	conn      *net.UDPConn
	isRunning bool
}

func New() (*IceCream, error) {
	iceCream := &IceCream{
		isRunning: false,
	}

	iceCream.init()

	return iceCream, nil
}

func (self *IceCream) listen() {
	for self.isRunning {
		buffer, _ := memorypool.GetInstance().Alloc(READ_BUFFER_SIZE)
		readLen, targetAddr, err := self.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalln(err)
		} else {
			f := func() {
				self.conn.WriteToUDP(buffer[:readLen], targetAddr)

				memorypool.GetInstance().Free(buffer)
			}

			threadpool.GetInstance().Start(f)
		}
	}
}

func (self *IceCream) init() {
}

func (self *IceCream) Start(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	self.udpAddr = udpAddr
	self.conn = conn
	self.isRunning = true

	self.listen()

	return nil
}

func (*IceCream) Stop() error {
	return nil
}
