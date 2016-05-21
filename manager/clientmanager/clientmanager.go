package clientmanager

import (
	"github.com/RexGene/icecream/icinterface"
	"log"
	"math/rand"
)

var instance *ClientManager

type ClientManager struct {
	dataMap map[uint32]icinterface.IClient
}

func GetInstance() *ClientManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func (self *ClientManager) GetClient(token uint32) icinterface.IClient {
	client, _ := self.dataMap[token]
	return client
}

func (self *ClientManager) AddClient(client icinterface.IClient) bool {
	token, isSuccess := self.MakeToken()
	if !isSuccess {
		log.Println("[-]token not empty")
		return false
	}

	client.SetToken(token)
	self.dataMap[token] = client
	return true
}

func (self *ClientManager) MakeToken() (uint32, bool) {
	count := 10
	for i := 0; i < count; i++ {
		token := rand.Uint32()
		if self.dataMap[token] == nil {
			return token, true
		}
	}

	return 0, false
}

func New() *ClientManager {
	return &ClientManager{
		dataMap: make(map[uint32]icinterface.IClient),
	}
}
