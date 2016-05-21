package handlermanager

import (
	"github.com/golang/protobuf/proto"
)

type HandlerManager struct {
	handlerMap map[uint32]func(proto.Message)
}

var instance *HandlerManager

func New() *HandlerManager {
	return &HandlerManager{
		handlerMap: make(map[uint32]func(proto.Message)),
	}
}

func GetInstance() *HandlerManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func (self *HandlerManager) RegistProtocol(id uint32, handleFunc func(proto.Message)) {
	self.handlerMap[id] = handleFunc
}

func (self *HandlerManager) HandleMessage(id uint32, msg proto.Message) {
	handleFunc := self.handlerMap[id]
	if handleFunc != nil {
		handleFunc(msg)
	}
}
