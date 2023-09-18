package utils

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// ForwarderMannager control all forwadrder manager
type ForwarderMannager struct {
	id          string // name of forwader audio or video
	forwadrders cmap.ConcurrentMap[string, *Forwarder]
	isClosed    bool
	logger      Log // init logger
}

// NewForwarderMannager create audio or video forwader
func NewForwarderMannager(id string, logger Log) *ForwarderMannager {
	f := &ForwarderMannager{
		id:          id,
		forwadrders: cmap.New[*Forwarder](),
		logger:      logger,
		isClosed:    false,
	}

	return f
}

// GetClient linter
func (f *ForwarderMannager) GetClient(trackID, pcID string) chan *Wrapper {
	fwd := f.getForwarder(trackID)
	if fwd == nil {
		return nil
	}
	c := fwd.getClient(pcID)
	if c != nil {
		return c.chann
	}
	return nil
}

// GetKeys return id of all forwarder
func (f *ForwarderMannager) GetKeys() []string {
	return f.forwadrders.Keys()
}

// Close lstringer
func (f *ForwarderMannager) Close() {
	f.setClose(true)
	f.closeForwaders()
}

// GetForwarder get forwarder of this id is exist or not
func (f *ForwarderMannager) GetForwarder(trackID string) *Forwarder {
	return f.getForwarder(trackID)
}

func (f *ForwarderMannager) AddNewForwarder(
	trackID string,
	codec string,
	handleChangeSSRC func(trackID string, pcIDs []string) error,
) *Forwarder {
	if oldFwd := f.getForwarder(trackID); oldFwd != nil {
		f.closeForwarder(trackID)
		return nil
	}
	// create new
	newForwader := NewForwarder(trackID, codec, f.logger, handleChangeSSRC)
	f.setForwarder(trackID, newForwader)
	f.logger.INFO(fmt.Sprintf("Add New %s forwarder successful", trackID), nil)
	return newForwader
}

// RemoveForwarder remove forwader with id
func (f *ForwarderMannager) RemoveForwarder(trackID string) {
	f.closeForwarder(trackID)
}

// Push to wrapper to specific id
func (f *ForwarderMannager) Push(trackID string, wrapper *Wrapper) {
	fwd := f.getForwarder(trackID)
	if fwd != nil {
		fwd.Push(wrapper)
	}
}

// Unregister unregis clientId to specific forwarder
func (f *ForwarderMannager) Unregister(trackID, pcID string) {
	fwd := f.getForwarder(trackID)
	if fwd != nil {
		fwd.UnRegister(pcID)
	}
}

// UnregisterAll linter
func (f *ForwarderMannager) UnregisterAll(pcID string) {
	for _, fwd := range f.forwadrders.Items() {
		fwd.UnRegister(pcID)
	}
}

// Register regis a client id to specific forwarder
func (f *ForwarderMannager) Register(trackID string, clientID string, handler func(trackID string, wrapper *Wrapper) error) {
	fwd := f.getForwarder(trackID)
	if fwd != nil {
		fwd.Register(&clientID, handler)
	}
}

func (f *ForwarderMannager) SendKeyframe(trackID, pcID string) error {
	fwd := f.getForwarder(trackID)
	if fwd == nil {
		f.logger.WARN(trackID+" fwd is nil", nil)
		return nil
	}
	fwd.SendKeyFrame(pcID)
	return nil
}

func (f *ForwarderMannager) SendAllKeyframe(pcID string) error {
	for _, fwd := range f.forwadrders.Items() {
		fwd.SendKeyFrame(pcID)
		f.logger.WARN(fmt.Sprintf("Send keyframe to %s_%s", f.id, pcID), nil)
	}
	return nil
}
