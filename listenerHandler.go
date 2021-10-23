package mdnsforwarder

import "net"

func NewListenerHandler(addr *net.UDPAddr) (*ListenerHandler, error) {
	handler := &ListenerHandler{
		listener: addr,
	}
	err := handler.Start()
	return handler, err
}

type ListenerHandler struct {
	listener *net.UDPAddr
	conn     *net.UDPConn
}

func (handler *ListenerHandler) Start() error {
	conn, err := net.ListenUDP("udp", handler.listener)
	if err != nil {
		return err
	}
	handler.conn = conn
	return nil
}

func (handler *ListenerHandler) Run(channel chan *RemoteUDPMessage) error {
	buffer := make([]byte, maxDatagramSize)

	for {
		numBytes, remoteAddr, err := handler.conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		if numBytes > 0 {
			trimmedData := buffer[:numBytes]
			message := &RemoteUDPMessage{
				Address:  remoteAddr,
				NumBytes: numBytes,
				Data:     trimmedData,
			}
			channel <- message
		}
	}
}

func (handler *ListenerHandler) Close() error {
	if handler.conn == nil {
		return nil
	}
	return handler.conn.Close()
}
