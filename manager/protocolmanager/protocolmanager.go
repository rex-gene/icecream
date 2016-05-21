package protocolmanager

import (
	"github.com/golang/protobuf/proto"
)

type ProtocolManager struct {
	protocolMap map[uint32]func() proto.Message
}

var instance *ProtocolManager

func New() *ProtocolManager {
	return &ProtocolManager{
		protocolMap: make(map[uint32]func() proto.Message),
	}
}

func GetInstance() *ProtocolManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func (self *ProtocolManager) RegistProtocol(id uint32, makeFunc func() proto.Message) {
	self.protocolMap[id] = makeFunc
}

func (self *ProtocolManager) GetProtocol(id uint32) proto.Message {
	makeFunc := self.protocolMap[id]
	if makeFunc != nil {
		return makeFunc()
	}

	return nil
}
