package databackupmanager

import (
	"errors"
	"github.com/RexGene/common/memorypool"
	"log"
)

var instance *DataBackupManager

const DEFAULT_CAP = 256

type DataBackupNode struct {
	Seq  uint16
	Data []byte
}

type DataBackupManager struct {
	data map[uint32][]DataBackupNode
}

func New() *DataBackupManager {
	return &DataBackupManager{
		data: make(map[uint32][]DataBackupNode),
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

func (self *DataBackupManager) Insert(token uint32, seq uint16, inputData []byte) {
	list := self.data[token]
	if list == nil {
		list = make([]DataBackupNode, 0, DEFAULT_CAP)
	}

	databackNode := DataBackupNode{
		Seq:  seq,
		Data: inputData,
	}

	self.data[token] = append(list, databackNode)
}

func (self *DataBackupManager) Clear() {
	for k, list := range self.data {
		for _, v := range list {
			self.FreeBuffer(v.Data)
		}
		delete(self.data, k)
	}
}

func (self *DataBackupManager) Remove(token uint32) error {
	list := self.data[token]
	if list == nil {
		return errors.New("token not found:" + string(token))
	}

	for _, v := range list {
		self.FreeBuffer(v.Data)
	}

	delete(self.data, token)
	return nil
}

func (self *DataBackupManager) GetDataList(token uint32) []DataBackupNode {
	return self.data[token]
}

func (self *DataBackupManager) FindAndRemove(token uint32, seq uint16) bool {
	list := self.data[token]
	if list == nil {
		log.Println("[!]token:", token, " seq:", seq, " drop!")
		return false
	}

	if len(list) == 0 {
		log.Println("[!]token:", token, " seq:", seq, " drop!")
		return false
	}

	index := 0
	for _, v := range list {
		index++
		if seq == v.Seq {
			break
		}
	}

	removeList := list[0:index]
	for _, v := range removeList {
		self.FreeBuffer(v.Data)
	}

	self.data[token] = list[index:]
	return true
}
