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

func SendData(cli icinterface.ISocket, buffer []byte, flag byte) {
	head := (*protocol.ICHead)(unsafe.Pointer(&buffer[0]))
	head.Flag = flag
	head.SrcSeqId = cli.GetSrcSeq()
	head.DstSeqId = cli.GetDstSeq()
	head.Token = 0
	head.CmdId = 0
	head.Sum = 0
	head.Len = uint16(len(buffer))
	head.Sum = GetSum(buffer)
	head.Token = cli.GetToken()

	log.Println("[?]send data:", *head)

	isNeedBackup := true
	if flag == protocol.ACK_FLAG || flag == protocol.RESET_FLAG {
		isNeedBackup = false
	}

	cli.SendData(buffer, isNeedBackup)
}

func SendMessage(
	socket icinterface.ISocket, msg proto.Message) {
	buffer := MakeBuffer(SEND_BUFFER_SIZE)

	msgData, err := proto.Marshal(msg)
	if err != nil {
		log.Println("[-]protocol marshal error")
		return
	}

	ptr := buffer[ICHEAD_SIZE:]
	for i, v := range msgData {
		ptr[i] = v
	}

	SendData(socket, buffer[:ICHEAD_SIZE+len(msgData)], protocol.PUSH_FLAG)
}

func CheckSum(buffer []byte) *protocol.ICHead {
	head := (*protocol.ICHead)(unsafe.Pointer(&buffer[0]))
	sum := head.Sum
	token := head.Token

	head.Sum = 0
	head.Token = 0

	sumValue := byte(0)
	for _, v := range buffer {
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

func HandlePacket(
	sender *datasendmanager.DataSendManager,
	tokenManager icinterface.ITokenManager,
	dataBackupManager *databackupmanager.DataBackupManager,
	addr *net.UDPAddr, buffer []byte, sock *socket.Socket) {

	head := CheckSum(buffer)
	if head == nil {
		log.Println("[!]check sum invaild")
		return
	}

	log.Println("[?]flag:", head.Flag)

	if head.Flag&protocol.ACK_FLAG != 0 {
		if head.Flag&protocol.START_FLAG == 0 {
			log.Println("[?]on ack")
			token := head.Token
			dataBackupManager.SendCmd(token, head.DstSeqId, nil, databackupmanager.FIND_AND_REMOVE)
		} else {
			log.Println("[?]on start ack")
			dataBackupManager.SendCmd(0, head.DstSeqId, nil, databackupmanager.FIND_AND_REMOVE)

			if sock != nil {
				sock.Format(head, addr, sender)
				tokenManager.AddSocketByToken(sock, head.Token)

				buffer := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
				SendData(sock, buffer, protocol.ACK_FLAG)
			}
		}
		return
	}

	if head.Flag&protocol.START_FLAG != 0 {
		log.Println("[?]on start")
		cli := socket.New()
		cli.Format(head, addr, sender)
		tokenManager.AddSocket(cli)

		buffer := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buffer, protocol.ACK_FLAG|protocol.START_FLAG)
		return
	}

	if head.Flag&protocol.RESET_FLAG != 0 {
		log.Println("[?]on reset")
		cli := tokenManager.GetSocket(head.Token)
		srcSeq := cli.GetSrcSeq()
		if head.DstSeqId > srcSeq {
			cli.SetSrcSeq(head.DstSeqId)
			dataBackupManager.SendCmd(head.Token, head.DstSeqId-1, nil, databackupmanager.FIND_AND_REMOVE)
		}
		return
	}

	if head.Flag&protocol.PUSH_FLAG == 0 {
		log.Println("[!]data flag is invalid:", head.Flag)
		return
	}

	log.Println("[?]on push")

	cli := tokenManager.GetSocket(head.Token)
	dstSeq := cli.GetDstSeq()
	if head.SrcSeqId < dstSeq {
		buffer := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buffer, protocol.RESET_FLAG)

	} else if head.SrcSeqId == dstSeq {
		buffer := dataBackupManager.MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buffer, protocol.ACK_FLAG)

		cli.IncDstSeq()
	} else {
		log.Println("[!]invalid seqId drop it:", head.SrcSeqId)
		return
	}

	cmdId := head.CmdId
	if cmdId != 0 {
		msg := protocolmanager.GetInstance().GetProtocol(cmdId)
		if msg != nil {
			proto.Unmarshal(buffer[ICHEAD_SIZE:], msg)
		}

		handlermanager.GetInstance().HandleMessage(cmdId, cli, msg)
	}

	memorypool.GetInstance().Free(buffer)
}