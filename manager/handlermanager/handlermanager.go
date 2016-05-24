package handlermanager

import (
	"github.com/RexGene/icecream/icinterface"
	"github.com/golang/protobuf/proto"
)

type HandlerManager struct {
	handlerMap map[uint32]func(icinterface.ISocket, proto.Message)
}

var instance *HandlerManager

func New() *HandlerManager {
	return &HandlerManager{
		handlerMap: make(map[uint32]func(icinterface.ISocket, proto.Message)),
	}
}

func GetInstance() *HandlerManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func (self *HandlerManager) RegistHandler(id uint32, handleFunc func(icinterface.ISocket, proto.Message)) {
	self.handlerMap[id] = handleFunc
}

func (self *HandlerManager) HandleMessage(id uint32, socket icinterface.ISocket, msg proto.Message) {
	handleFunc := self.handlerMap[id]
	if handleFunc != nil {
		handleFunc(socket, msg)
	}
}
