package datasendmanager

import (
	"github.com/RexGene/icecream/icinterface"
	"github.com/RexGene/icecream/manager/clientmanager"
	"github.com/RexGene/icecream/manager/databackupmanager"
	"net"
	"time"
)

const CMD_BUFFER_SIZE = 1024

const (
	SEND_DATA = iota
)

type CmdData struct {
	ClientData icinterface.IClient
	Data       []byte
	Option     int
}

type DataSendManager struct {
	cmdList chan CmdData
	conn    *net.UDPConn
}

var instance *DataSendManager

func (self *DataSendManager) Init(conn *net.UDPConn) {
	self.conn = conn
}

func GetInstance() *DataSendManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func New() *DataSendManager {
	return &DataSendManager{
		cmdList: make(chan CmdData, CMD_BUFFER_SIZE),
	}
}

func (self *DataSendManager) Execute() {
	dataBackupManager := databackupmanager.GetInstance()
	clientManager := clientmanager.GetInstance()
	conn := self.conn

	for {
		select {
		case cmd := <-self.cmdList:
			token := cmd.ClientData.GetToken()
			dataList := dataBackupManager.GetDataList(token)
			if dataList != nil {
				for _, backupData := range dataList {
					client := clientManager.GetClient(token)
					if client != nil {
						conn.WriteToUDP(backupData.Data, client.GetAddr())
					}
				}
			}

		case <-time.After(time.Second):
			dataMap := dataBackupManager.GetData()
			for token, dataList := range dataMap {
				if dataList != nil {
					for _, backupData := range dataList {
						client := clientManager.GetClient(token)
						if client != nil {
							conn.WriteToUDP(backupData.Data, client.GetAddr())
						}
					}
				}
			}
		}
	}
}

func (self *DataSendManager) SendCmd(client icinterface.IClient, data []byte, option int) {
	cmdData := CmdData{
		ClientData: client,
		Data:       data,
		Option:     option,
	}

	self.cmdList <- cmdData
}

func (self *DataSendManager) SendData(client icinterface.IClient, data []byte) {
	self.SendCmd(client, data, SEND_DATA)
	databackupmanager.GetInstance().SendCmd(client.GetToken(), client.GetSrcSeq(), data, databackupmanager.INSERT)
}
