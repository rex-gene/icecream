package icinterface

type ITokenManager interface {
	GetSocket(token uint32) ISocket
	AddSocketByToken(socket ISocket, token uint32)
	AddSocket(socket ISocket) bool
}
