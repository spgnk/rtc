package utils

// Fwdm linter
type Fwdm interface {
	UnregisterAll(peerConnectionID string) // unregister of fwd with input peer connection id
	// RegisterAll(clientID string, handler func(trackID string, wrapper *Wrapper) error)
	Register(fwdID string, clientID string, handler func(trackID string, wrapper *Wrapper) error)
	Unregister(trackID, pcID *string)
	AddNewForwarder(id, codec string) *Forwarder
	RemoveForwarder(id string)
	GetForwarder(id string) *Forwarder
	Push(id string, wrapper *Wrapper)
	GetKeys() []string
	Close()
	GetClient(trackID, pcID *string) chan *Wrapper
	GetLastTimeReceive() map[string]int64
	GetLastTimeReceiveBy(trackID string) int64
}
