package databackupmanager

import (
	"errors"
	"github.com/RexGene/common/memorypool"
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
	Seq  uint16
	Data []byte
	Size uint
}

type DataNode struct {
	sync.RWMutex
	Nodes []DataBackupNode
}

type DataBackupManager struct {
	data             map[uint32]*DataNode
	controlEventList chan ControlData
	exitEvent        chan bool
}

func New() *DataBackupManager {
	return &DataBackupManager{
		data:             make(map[uint32]*DataNode),
		controlEventList: make(chan ControlData, CONTROL_EVENT_LIST),
		exitEvent:        make(chan bool, 1),
	}
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
		node.Nodes = make([]DataBackupNode, 0, DEFAULT_CAP)
		self.data[token] = node
	}

	node.Lock()
	defer node.Unlock()

	databackNode := DataBackupNode{
		Seq:  seq,
		Data: inputData,
		Size: size,
	}

	node.Nodes = append(node.Nodes, databackNode)
}

func (self *DataBackupManager) clear() {
	for k, node := range self.data {
		node.Lock()
		defer node.Unlock()

		list := node.Nodes

		for _, v := range list {
			self.FreeBuffer(v.Data)
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
		self.FreeBuffer(v.Data)
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
	log.Println("[?]in")
	node := self.data[token]
	if node == nil {
		log.Println("[!]token:", token, " seq:", seq, " drop!", self.data)
		return false
	}

	node.Lock()
	defer node.Unlock()

	list := node.Nodes

	if len(list) == 0 {
		log.Println("[!]list is empty token:", token, " seq:", seq, " drop!")
		return false
	}

	index := 0
	for _, v := range list {
		index++
		if seq == v.Seq {
			log.Println("[?]found seq")
			break
		}
	}

	removeList := list[0:index]
	for _, v := range removeList {
		self.FreeBuffer(v.Data)
	}

	node.Nodes = list[index:]
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
				log.Println("[?]DataBackup insert token:", data.Token, " seq:", data.Seq)
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
