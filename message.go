package mdnsforwarder

import "net"

// UDPMessage is a message structure which represents
// one sent datagram from a listened message.
type UDPMessage struct {
	Interface *net.Interface
	Address   *net.UDPAddr
	NumBytes  int
	Data      []byte
}

// RemoteUDPMessage represents a data being sent to
// a unicast listener for forwarding the message to all
// local networks.
type RemoteUDPMessage struct {
	Address  *net.UDPAddr
	NumBytes int
	Data     []byte
}
