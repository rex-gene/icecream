package icecream

import (
	"github.com/RexGene/common/threadpool"
	"github.com/RexGene/icecream/manager/databackupmanager"
	"github.com/RexGene/icecream/manager/datasendmanager"
	"github.com/RexGene/icecream/net/converter"
	"github.com/RexGene/icecream/protocol"
	"log"
	"net"
	"unsafe"
)

const (
	MAX_CONNECT_COUNT = 1024
	READ_BUFFER_SIZE  = 65535
	ICHEAD_SIZE       = 16
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

func (self *IceCream) checkSum(buffer []byte) *protocol.ICHead {
	head := (*protocol.ICHead)(unsafe.Pointer(&buffer[0]))
	sum := head.Sum
	token := head.Token

	head.Sum = 0
	head.Token = 0

	sumValue := byte(0)
	for _, v := range buffer {
		sumValue ^= v
	}

	head.Token = token

	if len(buffer) != int(head.Len) {
		return nil
	}

	if sum != sumValue {
		return nil
	}

	return head
}

func (self *IceCream) listen() {
	for self.isRunning {
		buffer := converter.MakeBuffer(READ_BUFFER_SIZE)
		readLen, targetAddr, err := self.conn.ReadFromUDP(buffer)
		if err == nil {
			if readLen >= ICHEAD_SIZE {
				task := func() {
					converter.HandlePacket(targetAddr, buffer)
				}

				threadpool.GetInstance().Start(task)
			} else {
				log.Println("[!]data len too short:", readLen)
			}
		} else {
			log.Fatalln(err)
		}
	}
}

func (self *IceCream) init() {
}

func (self *IceCream) handleLogic() {
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

	dataSendManager := datasendmanager.GetInstance()
	dataSendManager.Init(conn)

	go dataSendManager.Execute()
	go databackupmanager.GetInstance().Execute()

	self.listen()

	return nil
}

func (*IceCream) Stop() error {
	return nil
}
