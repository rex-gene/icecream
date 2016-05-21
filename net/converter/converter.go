package converter

import (
	"github.com/RexGene/common/memorypool"
	"github.com/RexGene/icecream/icinterface"
	"github.com/RexGene/icecream/manager/clientmanager"
	"github.com/RexGene/icecream/manager/databackupmanager"
	"github.com/RexGene/icecream/manager/handlermanager"
	"github.com/RexGene/icecream/manager/protocolmanager"
	"github.com/RexGene/icecream/net/client"
	"github.com/RexGene/icecream/protocol"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"unsafe"
)

const ICHEAD_SIZE = 16

func GetSum(buffer []byte) byte {
	sumValue := byte(0)
	for _, v := range buffer {
		sumValue ^= v
	}

	return sumValue
}

func SendData(cli icinterface.IClient, buffer []byte, flag byte) {
	head := (*protocol.ICHead)(unsafe.Pointer(&buffer[0]))
	head.Flag = 0 | protocol.ACK_FLAG | protocol.PUSH_FLAG
	head.SrcSeqId = cli.GetSrcSeq()
	head.DstSeqId = cli.GetDstSeq()
	head.Token = 0
	head.CmdId = 0
	head.Sum = 0
	head.Len = uint16(len(buffer))
	head.Sum = GetSum(buffer)

	cli.SendData(buffer)
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

	head.Token = token

	if len(buffer) != int(head.Len) {
		return nil
	}

	if sum != sumValue {
		return nil
	}

	return head
}

func MakeBuffer(size uint) []byte {
	buffer, _ := memorypool.GetInstance().Alloc(size)
	return buffer
}

func HandlePacket(addr *net.UDPAddr, buffer []byte) {
	dataBackupManager := databackupmanager.GetInstance()

	head := CheckSum(buffer)
	if head == nil {
		log.Println("[!]check sum invaild")
		return
	}

	if head.Flag&protocol.PUSH_FLAG != 0 {
		cli := clientmanager.GetInstance().GetClient(head.Token)
		dstSeq := cli.GetDstSeq()
		if head.SrcSeqId < dstSeq {
			buffer := databackupmanager.GetInstance().MakeBuffer(ICHEAD_SIZE)
			SendData(cli, buffer, protocol.RESET_FLAG)

		} else if head.SrcSeqId == dstSeq {
			buffer := databackupmanager.GetInstance().MakeBuffer(ICHEAD_SIZE)
			SendData(cli, buffer, protocol.ACK_FLAG)

			cli.IncDstSeq()

		}
	} else if head.Flag&protocol.START_FLAG != 0 {
		cli := client.New()
		cli.Format(head, addr)
		clientmanager.GetInstance().AddClient(cli)

		buffer := databackupmanager.GetInstance().MakeBuffer(ICHEAD_SIZE)
		SendData(cli, buffer, protocol.ACK_FLAG|protocol.PUSH_FLAG)
	} else if head.Flag&protocol.RESET_FLAG != 0 {
		cli := clientmanager.GetInstance().GetClient(head.Token)
		srcSeq := cli.GetSrcSeq()
		if head.DstSeqId > srcSeq {
			cli.SetSrcSeq(head.DstSeqId)
			dataBackupManager.SendCmd(head.Token, head.DstSeqId-1, nil, databackupmanager.FIND_AND_REMOVE)
		}
	}

	if head.Flag&protocol.ACK_FLAG != 0 {
		dataBackupManager.SendCmd(head.Token, head.DstSeqId, nil, databackupmanager.FIND_AND_REMOVE)
	}

	cmdId := head.CmdId
	if cmdId != 0 {
		msg := protocolmanager.GetInstance().GetProtocol(cmdId)
		if msg != nil {
			proto.Unmarshal(buffer[ICHEAD_SIZE:], msg)
		}

		handlermanager.GetInstance().HandleMessage(cmdId, msg)
	}

	memorypool.GetInstance().Free(buffer)
}
