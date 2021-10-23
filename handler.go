package mdnsforwarder

import (
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func NewHandler(forwarder Forwarder) *Handler {
	return &Handler{
		forwarder: forwarder,
	}
}

type InterfaceCache struct {
	interfaceNetworkLookup map[string][]*net.IPNet
}

func (interfaceCache *InterfaceCache) IsInNetwork(networkInterface *net.Interface, address net.IP) bool {
	if interfaceCache.interfaceNetworkLookup == nil {
		interfaceCache.buildLookup()
	}

	networks, ok := interfaceCache.interfaceNetworkLookup[networkInterface.Name]
	if !ok {
		return false
	}

	for _, network := range networks {
		if network.Contains(address) {
			return true
		}
	}

	return false
}

func (interfaceCache *InterfaceCache) buildLookup() error {
	interfaceCache.interfaceNetworkLookup = map[string][]*net.IPNet{}
	lookup := interfaceCache.interfaceNetworkLookup
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, networkInterface := range interfaces {
		ipNetList := []*net.IPNet{}
		addrs, err := networkInterface.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			_, ipNet, err := net.ParseCIDR(addr.String())
			if err != nil {
				return err
			}
			ipNetList = append(ipNetList, ipNet)
		}

		lookup[networkInterface.Name] = ipNetList
	}

	return nil
}

type TimedInterfaceCache struct {
	interfaceCache *InterfaceCache
	createdAt      time.Time
	expires        time.Time
}

func (interfaceCache *TimedInterfaceCache) IsInNetwork(networkInterface *net.Interface, address net.IP) bool {
	return interfaceCache.interfaceCache.IsInNetwork(networkInterface, address)
}

func (interfaceCache *TimedInterfaceCache) Expired() bool {
	return time.Now().After(interfaceCache.expires)
}

func NewTimedInterfaceCache() *TimedInterfaceCache {
	return NewTimedInterfaceCacheExpiresAt(time.Now().Add(1 * time.Minute))
}

func NewTimedInterfaceCacheExpiresAt(expires time.Time) *TimedInterfaceCache {
	innerCache := &InterfaceCache{}
	return &TimedInterfaceCache{
		interfaceCache: innerCache,
		createdAt:      time.Now(),
		expires:        expires,
	}
}

// Handler includes the logic for
// forwarding and handling mdns
// messages.
type Handler struct {
	forwarder      Forwarder
	interfaces     map[*net.Interface]*InterfaceHandler
	targets        map[*net.UDPAddr]*TargetHandler
	listeners      map[*net.UDPAddr]*ListenerHandler
	interfaceCache *TimedInterfaceCache
}

func (handler *Handler) Start() error {
	handler.interfaces = map[*net.Interface]*InterfaceHandler{}
	handler.targets = map[*net.UDPAddr]*TargetHandler{}
	handler.listeners = map[*net.UDPAddr]*ListenerHandler{}

	for _, networkInterface := range handler.forwarder.MDNSNetworkInterfaces() {
		interfaceHandler, err := NewInterfaceHandler(networkInterface)
		if err != nil {
			return err
		}
		handler.interfaces[networkInterface] = interfaceHandler
	}

	for _, target := range handler.forwarder.GetTargets() {
		targetHandler := NewTargetHandler(target)
		handler.targets[target] = targetHandler
	}

	for _, listener := range handler.forwarder.ListenerIps() {
		listenerHandler, err := NewListenerHandler(listener)
		if err != nil {
			return err
		}

		handler.listeners[listener] = listenerHandler
	}

	return nil
}

func (handler *Handler) Run() error {
	messageChannel := make(chan *UDPMessage)
	for _, interfaceHandler := range handler.interfaces {
		go interfaceHandler.Run(messageChannel)
	}
	remoteMessageChannel := make(chan *RemoteUDPMessage)
	for _, listenerHandler := range handler.listeners {
		go listenerHandler.Run(remoteMessageChannel)
	}

	for {
		select {
		case udpMessage := <-messageChannel:
			log.Infof("Got local udp message at %s from %s", udpMessage.Interface.Name, udpMessage.Address.String())
			log.Debugf("Message includes: %s", udpMessage.Data)
			err := handler.handleLocalMessage(udpMessage)
			if err != nil {
				return err
			}
		case remoteUdpMessage := <-remoteMessageChannel:
			log.Infof("Got remote udp message from %s", remoteUdpMessage.Address.String())
			log.Tracef("Message includes: %s", remoteUdpMessage.Data)
			err := handler.handleRemoteMessage(remoteUdpMessage)
			if err != nil {
				return err
			}
		}
	}
}

func (handler *Handler) handleLocalMessage(udpMessage *UDPMessage) error {
	if handler.isOwnIP(udpMessage.Address.IP) && handler.forwarder.SkipOwnIP() {
		return nil
	}
	if handler.interfaceCache == nil || handler.interfaceCache.Expired() {
		handler.interfaceCache = NewTimedInterfaceCache()
	}
	interfaceCache := handler.interfaceCache

	for networkInterface, interfaceHandler := range handler.interfaces {
		if !interfaceCache.IsInNetwork(networkInterface, udpMessage.Address.IP) {
			log.Debugf("Moving message from %s to %s", udpMessage.Address.String(), networkInterface.Name)
			err := interfaceHandler.Send(udpMessage.Data)
			if err != nil {
				return err
			}
		}
	}

	for target, targetHandler := range handler.targets {
		log.Debugf("Forwarding local message from %s to remote target %s:%d", udpMessage.Address.String(), target.IP.String(), target.Port)
		err := targetHandler.Send(udpMessage.Data)
		if err != nil {
			log.Warnf("Couldn't send message to remote %s: %s", target.String(), err.Error())
		}
	}
	return nil
}

func (handler *Handler) isOwnIP(ip net.IP) bool {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, address := range addresses {
		items := strings.Split(address.String(), "/")
		if items[0] == ip.String() {
			return true
		}
	}

	return false
}

func (handler *Handler) handleRemoteMessage(udpMessage *RemoteUDPMessage) error {
	for networkInterface, interfaceHandler := range handler.interfaces {
		log.Debugf("Moving remote message from %s to %s", udpMessage.Address.String(), networkInterface.Name)
		err := interfaceHandler.Send(udpMessage.Data)
		if err != nil {
			return err
		}
	}
	return nil
}
