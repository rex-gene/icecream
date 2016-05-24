package datasendmanager

import (
	"github.com/RexGene/icecream/icinterface"
	"github.com/RexGene/icecream/manager/databackupmanager"
	"log"
	"net"
	"time"
)

const CMD_BUFFER_SIZE = 1024

const (
	SEND_DATA = iota
)

type CmdData struct {
	Socket icinterface.ISocket
	Data   []byte
	Option int
}

type DataSendManager struct {
	cmdList           chan CmdData
	exitEvent         chan bool
	conn              *net.UDPConn
	dataBackupManager *databackupmanager.DataBackupManager
	tokenManager      icinterface.ITokenManager
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
	}
}

func (self *DataSendManager) ExecuteForSocket(socket icinterface.ISocket) {
	dataBackupManager := self.dataBackupManager
	conn := self.conn

	for {
		select {
		case cmd := <-self.cmdList:
			_, err := conn.Write(cmd.Data)
			if err != nil {
				log.Println("[!]", err)
			}

		case <-self.exitEvent:
			return

		case <-time.After(time.Second):
			dataList := dataBackupManager.GetDataList(socket.GetToken())
			if dataList != nil {
				for _, backupData := range dataList {
					_, err := conn.Write(backupData.Data)
					if err != nil {
						log.Println("[!]", err)
					}
				}
			}

		}
	}
}

func (self *DataSendManager) Stop() {
	self.exitEvent <- true
}

func (self *DataSendManager) Execute() {
	dataBackupManager := self.dataBackupManager
	tokenManager := self.tokenManager
	conn := self.conn

	for {
		select {
		case cmd := <-self.cmdList:
			socket := cmd.Socket
			if socket != nil {
				conn.WriteToUDP(cmd.Data, socket.GetAddr())
			}
		case <-self.exitEvent:
			return
		case <-time.After(time.Second):
			dataMap := dataBackupManager.GetData()
			for token, dataList := range dataMap {
				if dataList != nil {
					for _, backupData := range dataList {
						socket := tokenManager.GetSocket(token)
						if socket != nil {
							conn.WriteToUDP(backupData.Data, socket.GetAddr())
						}
					}
				}
			}
		}
	}
}

func (self *DataSendManager) SendCmd(socket icinterface.ISocket, data []byte, option int) {
	cmdData := CmdData{
		Socket: socket,
		Data:   data,
		Option: option,
	}

	self.cmdList <- cmdData
}

func (self *DataSendManager) SendData(socket icinterface.ISocket, data []byte, isNeedBackup bool) {
	self.SendCmd(socket, data, SEND_DATA)
	if isNeedBackup {
		self.dataBackupManager.SendCmd(socket.GetToken(), socket.GetSrcSeq(), data, databackupmanager.INSERT)
		socket.IncSrcSeq()
	}
}
