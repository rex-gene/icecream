package connectormanager

import (
	"github.com/RexGene/icecream/net/connector"
	"sync"
)

var instance *ConnectorManager

type ConnectorManager struct {
	sync.RWMutex
	connectorMap map[string]*connector.Connector
}

func New() *ConnectorManager {
	return &ConnectorManager{
		connectorMap: make(map[string]*connector.Connector),
	}
}

func GetInstance() *ConnectorManager {
	if instance == nil {
		instance = New()
	}

	return instance
}

func (self *ConnectorManager) Insert(serverName string, connector *connector.Connector) {
	self.Lock()
	defer self.Unlock()

	self.connectorMap[serverName] = connector
}

func (self *ConnectorManager) Remove(serverName string) {
	self.Lock()
	defer self.Unlock()

	delete(self.connectorMap, serverName)
}

func (self *ConnectorManager) GetConnector(serverName string) *connector.Connector {
	self.RLock()
	defer self.RUnlock()

	return self.connectorMap[serverName]
}

func (self *ConnectorManager) Clear() {
	for key, connector := range self.connectorMap {
		connector.Stop()
		delete(self.connectorMap, key)
	}
}
