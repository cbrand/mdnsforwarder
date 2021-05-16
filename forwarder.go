package mdnsforwarder

import (
	"net"

	log "github.com/sirupsen/logrus"
)

// Forwarder is the main interface which is used
// to store information for the mdns forwarder to handle
// network multicasting.
type Forwarder interface {
	MDNSNetworkInterfaces() []*net.Interface
	ListenerIps() []*net.UDPAddr
	GetTargets() []*net.UDPAddr
	Run() error
}

func New(interfaces []*net.Interface, listenerIps []*net.UDPAddr, targets []*net.UDPAddr) Forwarder {
	return &ForwarderImpl{
		mdnsNetworkInterfaces: interfaces,
		listenerIps:           listenerIps,
		targets:               targets,
	}
}

type ForwarderImpl struct {
	mdnsNetworkInterfaces []*net.Interface
	listenerIps           []*net.UDPAddr
	targets               []*net.UDPAddr
}

func (forwarder *ForwarderImpl) MDNSNetworkInterfaces() []*net.Interface {
	return forwarder.mdnsNetworkInterfaces
}

func (forwarder *ForwarderImpl) ListenerIps() []*net.UDPAddr {
	return forwarder.listenerIps
}

func (forwarder *ForwarderImpl) GetTargets() []*net.UDPAddr {
	return forwarder.targets
}

func (forwarder *ForwarderImpl) Run() error {
	handler := NewHandler(forwarder)
	log.Debug("Bootstrapping mdnsforwarder")
	err := handler.Start()
	if err != nil {
		return err
	}

	log.Info("Running mdnsforwarder")
	return handler.Run()
}
