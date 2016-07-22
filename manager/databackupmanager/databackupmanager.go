package databackupmanager

import (
	"errors"
	"github.com/RexGene/common/memorypool"
	"github.com/RexGene/common/timermanager"
	"github.com/RexGene/common/timingwheel"
	"github.com/RexGene/icecream/icinterface"
	"log"
	"sync"
)

var instance *DataBackupManager

const DEFAULT_CAP = 256
const CONTROL_EVENT_LIST = 1024

const (
	INSERT = iota
	REMOVE
	FIND_AND_REMOVE
)

type ControlData struct {
	Token  uint32
	Seq    uint16
	Data   []byte
	Size   uint
	Option int
}

type DataBackupNode struct {
	sync.Mutex
	Data  []byte
	Size  uint
	Timer *timingwheel.BaseNode
	Count uint
}

type DataNode struct {
	sync.RWMutex
	Nodes map[uint16]*DataBackupNode
}

type DataBackupManager struct {
	bufferLock       sync.Mutex
	data             map[uint32]*DataNode
	controlEventList chan ControlData
	exitEvent        chan bool
	sender           icinterface.ISender
}

func New() *DataBackupManager {
	return &DataBackupManager{
		data:             make(map[uint32]*DataNode),
		controlEventList: make(chan ControlData, CONTROL_EVENT_LIST),
		exitEvent:        make(chan bool, 1),
	}
}

func (self *DataBackupManager) SetSender(sender icinterface.ISender) {
	self.sender = sender
}

func GetInstance() *DataBackupManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func (self *DataBackupManager) MakeBuffer(size uint) []byte {
	buf, _ := memorypool.GetInstance().Alloc(size)
	return buf
}

func (self *DataBackupManager) FreeBuffer(buffer []byte) {
	memorypool.GetInstance().Free(buffer)
}

func (self *DataBackupManager) insert(token uint32, seq uint16, inputData []byte, size uint) {
	node := self.data[token]
	if node == nil {
		node = new(DataNode)
		node.Nodes = make(map[uint16]*DataBackupNode)
		self.data[token] = node
	}

	node.Lock()
	defer node.Unlock()

	databackNode := &DataBackupNode{
		Data:  inputData,
		Size:  size,
		Count: 10,
	}

	log.Println("[?] databacknode:", databackNode.Data[:databackNode.Size], databackNode.Size)

	onTimeout := func() {
		databackNode.Lock()
		defer databackNode.Unlock()

		if !self.sender.Resend(uint(token), databackNode) {
			self.SendCmd(token, 0, nil, 0, REMOVE)
		}
	}

	databackNode.Timer = self.sender.AddTimer(onTimeout)

	node.Nodes[seq] = databackNode
}

func (self *DataBackupManager) clear() {
	for k, node := range self.data {
		node.Lock()
		defer node.Unlock()

		list := node.Nodes

		for _, v := range list {
			v.Lock()
			timermanager.RemoveTimer(v.Timer)
			self.FreeBuffer(v.Data)
			v.Data = nil
			v.Unlock()
		}

		delete(self.data, k)
	}
}

func (self *DataBackupManager) remove(token uint32) error {
	node := self.data[token]
	if node == nil {
		return errors.New("token not found:" + string(token))
	}

	node.Lock()
	defer node.Unlock()

	list := node.Nodes

	for _, v := range list {
		v.Lock()
		timermanager.RemoveTimer(v.Timer)
		self.FreeBuffer(v.Data)
		v.Data = nil
		v.Unlock()
	}

	delete(self.data, token)
	return nil
}

func (self *DataBackupManager) GetDataList(token uint32) *DataNode {
	return self.data[token]
}

func (self *DataBackupManager) GetData() map[uint32]*DataNode {
	return self.data
}

func (self *DataBackupManager) findAndRemove(token uint32, seq uint16) bool {
	node := self.data[token]
	if node == nil {
		log.Println("[!]token:", token, " seq:", seq, " drop!", self.data)
		return false
	}

	node.Lock()
	defer node.Unlock()

	nodes := node.Nodes
	dataNode := nodes[seq]
	if dataNode != nil {
		dataNode.Lock()
		defer dataNode.Unlock()

		timermanager.RemoveTimer(dataNode.Timer)
		self.FreeBuffer(dataNode.Data)
		dataNode.Data = nil
		delete(nodes, seq)
	}

	return true
}

func (self *DataBackupManager) SendCmd(token uint32, seq uint16, data []byte, size uint, option int) {
	cmd := ControlData{
		Token:  token,
		Seq:    seq,
		Data:   data,
		Size:   size,
		Option: option,
	}

	self.controlEventList <- cmd
}

func (self *DataBackupManager) Stop() {
	self.exitEvent <- true
}

func (self *DataBackupManager) Execute() {
	for {
		select {
		case data := <-self.controlEventList:
			switch data.Option {
			case INSERT:
				log.Println("[?]DataBackup insert token:", data.Token, " seq:", data.Seq, " size:", data.Size)
				self.insert(data.Token, data.Seq, data.Data, data.Size)
			case REMOVE:
				log.Println("[?]DataBackup remove", " token:", data.Token)
				self.remove(data.Token)
			case FIND_AND_REMOVE:
				log.Println("[?]DataBackup findAndRemove", " token:", data.Token, " seq:", data.Seq)
				self.findAndRemove(data.Token, data.Seq)
			}
		case <-self.exitEvent:
			self.clear()
			return
		}
	}
}
