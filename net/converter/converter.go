package converter

import (
	"github.com/RexGene/common/memorypool"
	"github.com/RexGene/icecream/icinterface"
	"github.com/RexGene/icecream/manager/databackupmanager"
	"github.com/RexGene/icecream/manager/datasendmanager"
	"github.com/RexGene/icecream/manager/handlermanager"
	"github.com/RexGene/icecream/manager/protocolmanager"
	"github.com/RexGene/icecream/net/socket"
	"github.com/RexGene/icecream/protocol"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"unsafe"
)

const ICHEAD_SIZE = 16
const SEND_BUFFER_SIZE = 65535
const SUM_FIX = 0x8C

func GetSum(buffer []byte) byte {
	sumValue := byte(0)
	for _, v := range buffer {
		sumValue ^= v
	}

	sumValue ^= SUM_FIX

	return sumValue
}

func SendData(cli icinterface.ISocket, buffer []byte, size uint, flag byte, cmdId int) {
	head := (*protocol.ICHead)(unsafe.Pointer(&buffer[0]))
	head.Flag = flag
	head.SrcSeqId = cli.GetSrcSeq()
	head.DstSeqId = cli.GetDstSeq()
	head.Token = 0
	head.CmdId = uint32(cmdId)
	head.Sum = 0
	head.Len = uint16(size)
	head.Sum = GetSum(buffer[:size])
	head.Token = cli.GetToken()

	log.Println("[?]send data:", size)

	isNeedBackup := true
	if flag == protocol.ACK_FLAG || flag == protocol.RESET_FLAG || flag == protocol.STOP_FLAG {
		isNeedBackup = false
	}

	cli.SendData(buffer, size, isNeedBackup)
}

func SendMessage(
	socket icinterface.ISocket, id int, msg proto.Message) {
	buffer := MakeBuffer(SEND_BUFFER_SIZE)

	msgData, err := proto.Marshal(msg)
	if err != nil {
		log.Println("[-]protocol marshal error")
		return
	}

	if len(msgData)+ICHEAD_SIZE > len(buffer) {
		log.Println("[-]data too long, can not be send")
		return
	}

	ptr := buffer[ICHEAD_SIZE:]
	for i, v := range msgData {
		ptr[i] = v
	}

	SendData(socket, buffer, uint(ICHEAD_SIZE+len(msgData)), protocol.PUSH_FLAG, id)
}

func CheckSum(buffer []byte) *protocol.ICHead {
	head := (*protocol.ICHead)(unsafe.Pointer(&buffer[0]))
	sum := head.Sum
	token := head.Token

	head.Sum = 0
	head.Token = 0

	buf := buffer[:head.Len]
	sumValue := byte(0)
	for _, v := range buf {
		sumValue ^= v
	}

	sumValue ^= SUM_FIX

	head.Token = token

	if sum != sumValue {
		log.Println("[!]check sum invaild: sum", sum, " sumValue:", sumValue)
		return nil
	}

	return head
}

func MakeBuffer(size uint) []byte {
	buffer, _ := memorypool.GetInstance().Alloc(size)
	return buffer
}

func FreeBuffer(buffer []byte) {
	memorypool.GetInstance().Free(buffer)
}

func SendStop(head *protocol.ICHead, sender *datasendmanager.DataSendManager,
	dataBackupManager *databackupmanager.DataBackupManager, addr *net.UDPAddr) {
	cli := socket.New()
	cli.Format(head, addr, sender)

	buffer := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
	SendData(cli, buffer, uint(len(buffer)), protocol.STOP_FLAG, 0)
}

func PushMessage(buffer []byte, cli icinterface.ISocket, handlerManager *handlermanager.HandlerManager) {
	head := (*protocol.ICHead)(unsafe.Pointer(&buffer[0]))
	cmdId := head.CmdId
	log.Println("[?]cmd:", cmdId)
	if cmdId != 0 {
		msg := protocolmanager.GetInstance().GetProtocol(cmdId)
		if msg != nil {
			proto.Unmarshal(buffer[ICHEAD_SIZE:], msg)
			handlerManager.PushMessage(cmdId, cli, msg)
		} else {
			log.Println("[!]protocol not found")
		}

	}

}

func HandlePacket(
	sender *datasendmanager.DataSendManager,
	tokenManager icinterface.ITokenManager,
	dataBackupManager *databackupmanager.DataBackupManager,
	handlerManager *handlermanager.HandlerManager,
	addr *net.UDPAddr, buffer []byte, sock *socket.Socket) bool {

	head := CheckSum(buffer)
	if head == nil {
		return true
	}

	log.Println("[?]flag:", head.Flag)

	if head.Flag&protocol.ACK_FLAG != 0 {
		if head.Flag&protocol.START_FLAG == 0 {
			log.Println("[?]on ack")
			token := head.Token
			dataBackupManager.SendCmd(token, head.DstSeqId, nil, 0, databackupmanager.FIND_AND_REMOVE)
		} else {
			log.Println("[?]on start ack")
			dataBackupManager.SendCmd(0, head.DstSeqId, nil, 0, databackupmanager.FIND_AND_REMOVE)

			if sock != nil {
				sock.Format(head, addr, sender)
				tokenManager.AddSocketByToken(sock, head.Token)

				buff := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
				SendData(sock, buff, uint(len(buff)), protocol.ACK_FLAG, 0)
				sock.IncDstSeq()
			}
		}
		return true
	}

	if head.Flag&protocol.START_FLAG != 0 {
		log.Println("[?]on start")
		cli := socket.New()
		cli.Format(head, addr, sender)
		tokenManager.AddSocket(cli)

		buff := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buff, uint(len(buff)), protocol.ACK_FLAG|protocol.START_FLAG, 0)
		cli.IncDstSeq()
		return true
	}

	if head.Flag&protocol.RESET_FLAG != 0 {
		log.Println("[!]on reset")
		cli := tokenManager.GetSocket(head.Token)
		if cli == nil {
			log.Println("[!]socket close")
			SendStop(head, sender, dataBackupManager, addr)
		} else {
			srcSeq := cli.GetSrcSeq()
			log.Println("[!]handle reset: head.DstSeqId:", head.DstSeqId, " srcSeq:", srcSeq)
			dataBackupManager.SendCmd(head.Token, head.DstSeqId-1, nil, 0, databackupmanager.FIND_AND_REMOVE)
		}
		return true
	}

	if head.Flag&protocol.STOP_FLAG != 0 {
		log.Println("[?]on stop")
		return true
	}

	if head.Flag&protocol.PUSH_FLAG == 0 {
		log.Println("[!]data flag is invalid:", head.Flag)
		return true
	}

	cli := tokenManager.GetSocket(head.Token)
	if cli == nil {
		log.Println("[!]send stop")
		SendStop(head, sender, dataBackupManager, addr)
		return true
	}

	log.Println("[?]on push")

	dstSeq := cli.GetDstSeq()
	if head.SrcSeqId == dstSeq {
		buff := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buff, uint(len(buff)), protocol.ACK_FLAG, 0)
		log.Println("[?]send ack:", head.SrcSeqId, dstSeq)

		var handleBackupList = func(data []byte) {
			PushMessage(buffer, cli, handlerManager)
			cli.IncDstSeq()
		}

		cli.IncDstSeq()
		cli.SetDstSeq(cli.EachBackupPacket(cli.GetDstSeq(), handleBackupList))
	} else if head.SrcSeqId-dstSeq < uint16(0x8000) {
		cli.InsertBackupList(head.SrcSeqId, buffer)

		buff := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buff, uint(len(buff)), protocol.ACK_FLAG, 0)
		log.Println("[?]send ack:", head.SrcSeqId, dstSeq)
		return false
	} else {

		buff := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buff, uint(len(buff)), protocol.ACK_FLAG, 0)
		log.Println("[?]send ack:", head.SrcSeqId, dstSeq)
		return true
	}

	cmdId := head.CmdId
	log.Println("[?]cmd:", cmdId)
	if cmdId != 0 {
		msg := protocolmanager.GetInstance().GetProtocol(cmdId)
		if msg != nil {
			proto.Unmarshal(buffer[ICHEAD_SIZE:], msg)
			handlerManager.PushMessage(cmdId, cli, msg)
		} else {
			log.Println("[!]protocol not found")
		}

	}

	return true
}
