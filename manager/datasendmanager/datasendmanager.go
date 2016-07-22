package datasendmanager

import (
	"github.com/RexGene/common/timermanager"
	"github.com/RexGene/common/timingwheel"
	"github.com/RexGene/icecream/icinterface"
	"github.com/RexGene/icecream/manager/databackupmanager"
	"log"
	"net"
	"time"
)

const CMD_BUFFER_SIZE = 1024

const (
	SEND_DATA = iota
	SEND_DATA_AND_FREE
)

type CmdData struct {
	Socket icinterface.ISocket
	Data   []byte
	Size   uint
	Option int
}

type DataSendManager struct {
	cmdList           chan CmdData
	exitEvent         chan bool
	conn              *net.UDPConn
	dataBackupManager *databackupmanager.DataBackupManager
	tokenManager      icinterface.ITokenManager
	timerManager      *timermanager.TimerManager
}

var instance *DataSendManager

func (self *DataSendManager) GetDataBackupManager() *databackupmanager.DataBackupManager {
	return self.dataBackupManager
}

func (self *DataSendManager) Init(conn *net.UDPConn, dataBackupManager *databackupmanager.DataBackupManager,
	tokenManager icinterface.ITokenManager) {
	self.conn = conn
	self.dataBackupManager = dataBackupManager
	self.tokenManager = tokenManager
}

func GetInstance() *DataSendManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func New() *DataSendManager {
	return &DataSendManager{
		cmdList:           make(chan CmdData, CMD_BUFFER_SIZE),
		exitEvent:         make(chan bool, 1),
		dataBackupManager: nil,
		timerManager:      timermanager.New(),
	}
}

func (self *DataSendManager) AddTimer(cb func()) *timingwheel.BaseNode {
	return self.timerManager.AddTimerForever(2, cb)
}

func (self *DataSendManager) ExecuteForSocket(socket icinterface.ISocket) {
	conn := self.conn
	tickTime := time.Millisecond * 100
	for {
		select {
		case cmd := <-self.cmdList:
			_, err := conn.Write(cmd.Data[:cmd.Size])
			if err != nil {
				log.Println("[!]", err)
			}

		case <-self.exitEvent:
			return

		case <-time.After(tickTime):
			self.timerManager.Tick()
		}
	}
}

func (self *DataSendManager) Stop() {
	self.exitEvent <- true
}

func (self *DataSendManager) Resend(token uint, itf interface{}) bool {
	conn := self.conn
	tokenManager := self.tokenManager
	socket := tokenManager.GetSocket(uint32(token))
	if socket != nil {
		backupData := itf.(*databackupmanager.DataBackupNode)
		if backupData.Count == 0 || backupData.Data == nil {
			return false
		}

		log.Println("[?] resend data:", backupData.Data[:backupData.Size], " size:", backupData.Size)
		backupData.Count--
		conn.WriteToUDP(backupData.Data[:backupData.Size], socket.GetAddr())

		return true
	} else {
		return false
	}
}

func (self *DataSendManager) Execute() {
	conn := self.conn
	tickTime := time.Millisecond * 100
	for {
		select {
		case cmd := <-self.cmdList:
			socket := cmd.Socket
			if socket != nil {
				conn.WriteToUDP(cmd.Data[:cmd.Size], socket.GetAddr())
				if cmd.Option == SEND_DATA_AND_FREE {
					self.dataBackupManager.FreeBuffer(cmd.Data)
				}
			}
		case <-self.exitEvent:
			return
		case <-time.After(tickTime):
			self.timerManager.Tick()
		}
	}
}

func (self *DataSendManager) SendCmd(socket icinterface.ISocket, data []byte, size uint, option int) {
	cmdData := CmdData{
		Socket: socket,
		Data:   data,
		Option: option,
		Size:   size,
	}

	self.cmdList <- cmdData
}

func (self *DataSendManager) SendData(socket icinterface.ISocket, data []byte, size uint, isNeedBackup bool) {
	if isNeedBackup {
		self.SendCmd(socket, data, size, SEND_DATA)
		self.dataBackupManager.SendCmd(socket.GetToken(), socket.GetSrcSeq(), data, size, databackupmanager.INSERT)
		socket.IncSrcSeq()
	} else {
		self.SendCmd(socket, data, size, SEND_DATA_AND_FREE)
	}
}
