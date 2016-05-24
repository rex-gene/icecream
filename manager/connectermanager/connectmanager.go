package connectermanager

import (
	"github.com/RexGene/icecream/net/connecter"
	"sync"
)

var instance *ConnecterManager

type ConnecterManager struct {
	sync.RWMutex
	connecterMap map[string]*connecter.Connecter
}

func New() *ConnecterManager {
	return &ConnecterManager{
		connecterMap: make(map[string]*connecter.Connecter),
	}
}

func GetInstance() *ConnecterManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func (self *ConnecterManager) Insert(serverName string, connecter *connecter.Connecter) {
	self.Lock()
	defer self.Unlock()

	self.connecterMap[serverName] = connecter
}

func (self *ConnecterManager) Remove(serverName string) {
	self.Lock()
	defer self.Unlock()

	delete(self.connecterMap, serverName)
}

func (self *ConnecterManager) GetConnecter(serverName string) *connecter.Connecter {
	self.RLock()
	defer self.RUnlock()

	return self.connecterMap[serverName]
}

func (self *ConnecterManager) Clear() {
	for key, connecter := range self.connecterMap {
		connecter.Stop()
		delete(self.connecterMap, key)
	}
}
