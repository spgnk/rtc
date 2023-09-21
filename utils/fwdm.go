package utils

import (
	"fmt"
	"sync"
)

const (
	unregister = "unregister"
	register   = "register"
)

// ClientDataTime linter
type ClientDataTime struct {
	id string
	t  int64
}

// FwdmAction linter
type FwdmAction struct {
	Action
	result chan *Forwarder // return fwd if exist
	pcID   *string
	codec  string
}

// ForwarderMannager control all forwadrder manager
type ForwarderMannager struct {
	id            string // name of forwader audio or video
	forwadrders   map[string]*Forwarder
	isClosed      bool
	msgChann      chan *FwdmAction // do what erver
	hub           chan *FwdmAction // dispatch all data
	dataTimeChann chan *ClientDataTime
	dataTime      map[string]int64
	mutex         sync.RWMutex
	logger        Log
	handleSSRC    func(trackID string, pcIDs []string, codec string)
}

// NewForwarderMannager create audio or video forwader
func NewForwarderMannager(
	id string,
	logger Log,
	handleSSRC func(trackID string, pcIDs []string, codec string),
) *ForwarderMannager {
	f := &ForwarderMannager{
		id:            id,
		forwadrders:   make(map[string]*Forwarder),
		msgChann:      make(chan *FwdmAction, maxChanSize),
		hub:           make(chan *FwdmAction, maxChanSize),
		dataTimeChann: make(chan *ClientDataTime, maxChanSize),
		dataTime:      make(map[string]int64),
		isClosed:      false,
		logger:        logger,
		handleSSRC:    handleSSRC,
	}

	go f.serve()
	go f.dispatch()
	go f.updateClientDataTime()
	return f
}

func (f *ForwarderMannager) updateClientDataTime() {
	var c *ClientDataTime
	var open bool
	for {
		c, open = <-f.dataTimeChann
		if !open || f.checkClose() {
			return
		}
		f.setDatatime(c)
		c = nil
	}
}

// Serve to run
func (f *ForwarderMannager) serve() {
	defer close(f.msgChann)
	for {
		action, open := <-f.msgChann
		if !open {
			return
		}
		f.choosing(action)
	}
}

// Close lstringer
func (f *ForwarderMannager) Close() {
	f.setClose(true)
	f.closeForwaders()
}

// GetClient linter
func (f *ForwarderMannager) GetClient(trackID, pcID *string) chan *Wrapper {
	fwd := f.getForwarder(trackID)
	if fwd == nil {
		f.logger.WARN("fwd is nil", map[string]any{
			"track_id": *trackID,
			"pc_id":    *pcID,
		})
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
	temp := make([]string, 0)
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for id := range f.forwadrders {
		temp = append(temp, id)
	}
	return temp
}

// GetForwarder get forwarder of this id is exist or not
func (f *ForwarderMannager) GetForwarder(id string) *Forwarder {
	return f.getForwarder(&id)
}

// AddNewForwarder linter
func (f *ForwarderMannager) AddNewForwarder(fwdID, codec string) *Forwarder {
	result := make(chan *Forwarder, 1)
	newAction := &FwdmAction{
		result: result,
		codec:  codec,
	}
	newAction.do = add
	newAction.id = &fwdID
	f.msgChann <- newAction
	fwd := <-result
	return fwd
}

// RemoveForwarder remove forwader with id
func (f *ForwarderMannager) RemoveForwarder(id string) {
	newAction := &FwdmAction{}
	newAction.id = &id
	newAction.do = closing
	f.msgChann <- newAction
}

// Push to wrapper to specific id
func (f *ForwarderMannager) Push(id string, wrapper *Wrapper) {
	newAction := &FwdmAction{}
	newAction.id = &id
	newAction.do = add
	newAction.data = wrapper
	f.msgChann <- newAction
}

// Unregister unregis clientId to specific forwarder
func (f *ForwarderMannager) Unregister(trackID, pcID *string) {
	newAction := &FwdmAction{}
	newAction.id = trackID
	newAction.pcID = pcID
	newAction.do = unregister
	f.msgChann <- newAction
}

// UnregisterAll linter
func (f *ForwarderMannager) UnregisterAll(peerConnectionID string) {
	// fwds := f.getAllFwd()
	// for _, fwd := range fwds {
	// 	fwd.UnRegister(&peerConnectionID)
	// }
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for _, fwd := range f.forwadrders {
		fwd.UnRegister(&peerConnectionID)
	}
}

// Register regis a client id to specific forwarder
func (f *ForwarderMannager) Register(trackID string, clientID string, handler func(trackID string, wrapper *Wrapper) error) {
	newAction := &FwdmAction{
		Action: Action{
			id: &trackID,
			client: &Client{
				handler: handler,
			},
		},
		pcID: &clientID,
	}
	newAction.do = register
	f.msgChann <- newAction
}

func (f *ForwarderMannager) choosing(action *FwdmAction) {
	switch action.do {
	case add:
		f.addNewForwarder(action.id, action.codec, action.result)
	// case closing:
	// go f.closeForwarder(action.id)
	case hub:
		f.hub <- action
	case unregister:
		go f.unregister(action.id, action.pcID)
	case register:
		go f.register(action)
	default:
		return
	}
}

func (f *ForwarderMannager) addNewForwarder(fwdID *string, codec string, result chan *Forwarder) {
	if oldFwd := f.getForwarder(fwdID); oldFwd != nil {
		result <- oldFwd
		return
	}
	// create new
	newForwader := NewForwarder(*fwdID, codec, f.dataTimeChann, f.logger, f.handleSSRC)
	f.setForwarder(fwdID, newForwader)
	f.logger.INFO(fmt.Sprintf("Add New %s forwarder successful", *fwdID), nil)
	result <- newForwader
}

func (f *ForwarderMannager) dispatch() {
	var msg *FwdmAction
	var open bool
	defer close(f.hub)
	for {
		msg, open = <-f.hub
		if !open {
			return
		}
		go f.forward(msg)
		msg = nil
	}
}

func (f *ForwarderMannager) forward(msg *FwdmAction) {
	forwarder := f.getForwarder(msg.id)
	if forwarder == nil {
		f.logger.WARN("forwarder is nil. Cannot push", map[string]any{
			"track_id": *msg.id,
		})
		return
	}
	forwarder.Push(msg.data)
}

func (f *ForwarderMannager) unregister(trackID, pcID *string) {
	if forwardfer := f.getForwarder(trackID); forwardfer != nil {
		forwardfer.closeClient(pcID)
	}
}

func (f *ForwarderMannager) register(action *FwdmAction) {
	if f.checkClose() {
		return
	}
	forwardfer := f.getForwarder(action.id)
	if forwardfer == nil {
		forwardfer = f.AddNewForwarder(*action.id, action.codec) // TODO check maybe stuck here
	}
	forwardfer.Register(action.pcID, action.client.handler)
}

// GetLastTimeReceive linter
func (f *ForwarderMannager) GetLastTimeReceive() map[string]int64 {
	temp := make(map[string]int64)

	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for k, v := range f.dataTime {
		temp[k] = v
	}

	return temp
}

// GetLastTimeReceiveBy linter
func (f *ForwarderMannager) GetLastTimeReceiveBy(trackID string) int64 {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.dataTime[trackID]
}

func (f *ForwarderMannager) SendKeyframe(trackID, pcID string) error {
	fwd := f.getForwarder(&trackID)
	if fwd == nil {
		f.logger.WARN("fwd is nil", map[string]any{
			"track_id": trackID,
			"pc_id":    pcID,
		})
		return nil
	}
	fwd.SendKeyFrame(&pcID)
	return nil
}

func (f *ForwarderMannager) SendAllKeyframe(pcID string) error {
	f.logger.WARN(fmt.Sprintf("Send All keyframe from %s to %s", f.id, pcID), nil)
	for _, fwd := range f.forwadrders {
		fwd.SendKeyFrame(&pcID)
		f.logger.WARN(fmt.Sprintf("Send keyframe to %s_%s", f.id, pcID), nil)
	}
	return nil
}
