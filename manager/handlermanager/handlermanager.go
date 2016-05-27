package handlermanager

import (
	"github.com/RexGene/icecream/icinterface"
	"github.com/golang/protobuf/proto"
)

const MAX_HANDLE_COUNT = 1024

type LogicMsg struct {
	cmdId uint32
	sock  icinterface.ISocket
	msg   proto.Message
}

type HandlerManager struct {
	handlerMap   map[uint32]func(icinterface.ISocket, proto.Message)
	logicMsgList chan LogicMsg
	exitEvent    chan bool
}

var instance *HandlerManager

func New() *HandlerManager {
	return &HandlerManager{
		handlerMap:   make(map[uint32]func(icinterface.ISocket, proto.Message)),
		logicMsgList: make(chan LogicMsg, MAX_HANDLE_COUNT),
		exitEvent:    make(chan bool, 1),
	}
}

func (self *HandlerManager) PushMessage(cmdId uint32, sock icinterface.ISocket, msg proto.Message) {
	logicMsg := LogicMsg{
		cmdId: cmdId,
		sock:  sock,
		msg:   msg,
	}

	self.logicMsgList <- logicMsg
}

func (self *HandlerManager) Execute() {
	for {
		select {
		case logicMsg := <-self.logicMsgList:
			self.HandleMessage(uint32(logicMsg.cmdId), logicMsg.sock, logicMsg.msg)
		case <-self.exitEvent:
			return
		}
	}
}

func (self *HandlerManager) Stop() {
	self.exitEvent <- true
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
