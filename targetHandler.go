package mdnsforwarder

import "net"

func NewTargetHandler(target *net.UDPAddr) *TargetHandler {
	return &TargetHandler{
		target: target,
	}
}

type TargetHandler struct {
	target *net.UDPAddr
	conn   net.Conn
}

func (handler *TargetHandler) Send(data []byte) error {
	if handler.conn == nil {
		connection, err := net.Dial("udp", handler.target.String())
		if err != nil {
			return err
		}
		handler.conn = connection
	}
	_, err := handler.conn.Write(data)
	return err
}
