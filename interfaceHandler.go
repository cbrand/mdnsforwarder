package mdnsforwarder

import (
	"context"
	"fmt"
	"net"
	"syscall"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/ipv4"
)

func NewInterfaceHandler(networkInterface *net.Interface) (*InterfaceHandler, error) {
	return NewInterfaceHandlerWithAddress(networkInterface, mdnsMulitcastAddress)
}

func NewInterfaceHandlerWithAddress(networkInterface *net.Interface, multicastAddress string) (*InterfaceHandler, error) {
	handler := &InterfaceHandler{
		networkInterface: networkInterface,
		address:          multicastAddress,
	}
	err := handler.Start()
	return handler, err
}

func reusePort(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		syscall.SetsockoptInt(int(descriptor), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
}

type InterfaceHandler struct {
	networkInterface *net.Interface
	address          string
	packetConnection *ipv4.PacketConn
	udpAddress       *net.UDPAddr
}

func (interfaceHandler *InterfaceHandler) Start() error {
	addrs, err := interfaceHandler.networkInterface.Addrs()
	if err != nil {
		return err
	}

	networkAddr := addrs[0]
	addr, _, err := net.ParseCIDR(networkAddr.String())
	if err != nil {
		return err
	}
	udpAddr, err := net.ResolveUDPAddr("udp", interfaceHandler.address)
	if err != nil {
		return err
	}
	interfaceHandler.udpAddress = udpAddr

	connAddr := fmt.Sprintf("%s:%d", addr.String(), udpAddr.Port)
	log.Debugf("Binding to UDP %s", connAddr)
	config := &net.ListenConfig{Control: reusePort}
	conn, err := config.ListenPacket(context.Background(), "udp", udpAddr.String())
	if err != nil {
		return err
	}
	group := udpAddr.IP

	packetConn := ipv4.NewPacketConn(conn)
	log.Tracef("Joining Multicast group %s", group)
	if err := packetConn.JoinGroup(interfaceHandler.networkInterface, &net.UDPAddr{IP: group}); err != nil {
		return err
	}

	if err := packetConn.SetControlMessage(ipv4.FlagDst, true); err != nil {
		return err
	}

	interfaceHandler.packetConnection = packetConn
	return nil
}

func (interfaceHandler *InterfaceHandler) Run(channel chan *UDPMessage) error {
	for {
		buffer := make([]byte, maxDatagramSize)
		numBytes, controlMessage, _, err := interfaceHandler.packetConnection.ReadFrom(buffer)
		if err != nil {
			return err
		}
		if controlMessage.Dst.IsMulticast() && controlMessage.Dst.Equal(interfaceHandler.udpAddress.IP) {
			if numBytes > 0 && !interfaceHandler.sendFromSelf(controlMessage.Src) {
				message := &UDPMessage{
					Interface: interfaceHandler.networkInterface,
					Address:   &net.UDPAddr{IP: controlMessage.Src, Port: interfaceHandler.udpAddress.Port, Zone: interfaceHandler.udpAddress.Zone},
					NumBytes:  numBytes,
					Data:      buffer,
				}
				channel <- message
			}
		}
	}
}

func (interfaceHandler *InterfaceHandler) Send(message []byte) error {
	packetConnection := interfaceHandler.packetConnection
	ctrl := &ipv4.ControlMessage{
		IfIndex: interfaceHandler.networkInterface.Index,
	}
	_, err := packetConnection.WriteTo(message, ctrl, interfaceHandler.udpAddress)
	return err
}

func (interfaceHandler *InterfaceHandler) Stop() error {
	return interfaceHandler.packetConnection.Close()
}

func (interfaceHandler *InterfaceHandler) sendFromSelf(ipAddress net.IP) bool {
	addrs, err := interfaceHandler.networkInterface.Addrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if addr.String() == ipAddress.String() {
			return true
		}
	}
	return false
}
