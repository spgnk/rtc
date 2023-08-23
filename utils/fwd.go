package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lamhai1401/gologs/logs"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

var wrapPool = sync.Pool{New: func() interface{} {
	return new(Wrapper)
}}

var pkgPool = sync.Pool{New: func() interface{} {
	return &rtp.Packet{}
}}

const (
	maxChanSize = 1024 * 2
	add         = "add"
	closing     = "closing"
	hub         = "hub"
)

// Client linter
type Client struct {
	chann      chan *Wrapper
	handler    func(trackID string, wrapper *Wrapper) error
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// Wrapper linter
type Wrapper struct {
	Duration time.Duration
	Pkg      *rtp.Packet // save rtp packet
	Data     []byte      `json:"rtp"`    // packet to write
	Kind     *string     `json:"kind"`   // audio or video
	SeatID   *int        `json:"seatID"` // stream id number 1-2-3-4
	Type     *string     `json:"type"`   // type off wrapper data - ok - ping - pong
}

// Action linter
type Action struct {
	do     string  // CRUD
	id     *string // for client id
	client *Client
	data   *Wrapper
}

// Forwarder linter
type Forwarder struct {
	id            string // stream id
	isClosed      bool
	clients       map[string]*Client // save all client with handler
	hub           chan *Wrapper      // dispatch all data
	msgChann      chan *Action       // do what erver
	ctx           context.Context
	cancelFunc    context.CancelFunc
	dataTimeChann chan *ClientDataTime
	mutex         sync.RWMutex
	keyframe      *rtp.Packet // save to keyframe
	keyframeChann chan []byte
	codec         string // vp8/vp9/h264
}

// NewForwarder return new forwarder
func NewForwarder(id string, codec string, dataTimeChann chan *ClientDataTime) *Forwarder {
	ctx, cancel := context.WithCancel(context.Background())
	f := &Forwarder{
		id:            id,
		hub:           make(chan *Wrapper, maxChanSize),
		msgChann:      make(chan *Action, maxChanSize),
		clients:       make(map[string]*Client),
		keyframe:      &rtp.Packet{},
		keyframeChann: make(chan []byte, maxChanSize),
		isClosed:      false,
		ctx:           ctx,
		cancelFunc:    cancel,
		dataTimeChann: dataTimeChann,
		codec:         codec,
	}

	go f.serve()
	go f.dispatch()
	go f.serveKeyFrame()

	return f
}

// func (f *Forwarder) getLastReceiveData() int64 {
// 	f.mutex.RLock()
// 	defer f.mutex.RUnlock()
// 	return f.lastReceiveData
// }

// func (f *Forwarder) setLastReceiveData(t int64) {
// 	f.mutex.Lock()
// 	defer f.mutex.Unlock()
// 	f.lastReceiveData = t
// }

// Serve to run
func (f *Forwarder) serve() {
	defer close(f.msgChann)
	for {
		select {
		case action := <-f.msgChann:
			f.choosing(action)
		case <-f.ctx.Done():
			return
		}
	}
}

func (f *Forwarder) dispatch() {
	var msg *Wrapper
	var open bool
	defer close(f.hub)
	for {
		select {
		case msg, open = <-f.hub:
			if !open {
				return
			}

			temp := make([]byte, len(msg.Data)) // add here to detect key frame
			copy(msg.Data, temp)
			f.handleKeyFrame(temp)

			f.forward(msg)
			// go f.setLastReceiveData(time.Now().UnixMilli())
			f.dataTimeChann <- &ClientDataTime{
				id: f.getID(),
				t:  time.Now().UnixMilli(),
			}
			msg = nil
		case <-f.ctx.Done():
			return
		}
	}
}

// Close to close all serve
func (f *Forwarder) Close() {
	if !f.checkClose() {
		f.setClose(true)
		f.cancelFunc()
		f.info(fmt.Sprintf("%s forwarder was closed", f.getID()))
	}
}

func (f *Forwarder) choosing(action *Action) {
	switch action.do {
	case add:
		f.addClient(action.id, action.client)
	case closing:
		f.closeClient(action.id)
	case hub:
		f.hub <- action.data
	default:
		return
	}
}

func (f *Forwarder) forward(wrapper *Wrapper) {
	f.mutex.RLock()
	defer func() {
		f.mutex.RUnlock()
		handlepanic(nil)
	}()
	for _, client := range f.clients {
		client.chann <- wrapper
	}
}

// RemoveClient linter
func (f *Forwarder) RemoveClient(clientID *string) {
	f.msgChann <- &Action{
		do: closing,
		id: clientID,
	}
}

// AddClient linter
func (f *Forwarder) AddClient(clientID *string, client *Client) {
	f.msgChann <- &Action{
		do:     add,
		client: client,
		id:     clientID,
	}
}

// Hub linter
func (f *Forwarder) Hub(wrapper *Wrapper) {
	f.msgChann <- &Action{
		do:   hub,
		data: wrapper,
	}
}

// Push new wrapper to server chan
func (f *Forwarder) Push(wrapper *Wrapper) {
	if f.checkClose() {
		f.info("fwd was closed")
		return
	}
	f.Hub(wrapper)
}

// UnRegister linter
func (f *Forwarder) UnRegister(clientID *string) {
	f.RemoveClient(clientID)
}

// Register linter
func (f *Forwarder) Register(clientID *string, handler func(trackID string, wrapper *Wrapper) error) {
	// remove client if exist
	// if f.getClient(clientID) != nil {
	// 	f.info(*clientID, " already exist. No need register")
	// 	return
	// }
	// f.RemoveClient(clientID)
	f.closeClient(clientID)

	ctx, cancel := context.WithCancel(context.Background())
	newClient := &Client{
		chann:      make(chan *Wrapper, maxChanSize),
		handler:    handler,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	f.AddClient(clientID, newClient)

	go f.collectData(clientID, newClient)
}

func (f *Forwarder) collectData(clientID *string, c *Client) {
	// defer func() {
	// 	// for pushing async data cause crash
	// 	ticker := time.NewTicker(5 * time.Second)
	// 	<-ticker.C
	// 	close(c.chann)
	// }()
	var err error
	var w *Wrapper
	var open bool
	for {
		select {
		case w, open = <-c.chann:
			if !open {
				return
			}
			buff := wrapPool.Get().(*Wrapper)
			pkg := pkgPool.Get().(*rtp.Packet)
			pkg.Unmarshal(w.Data)
			buff.Pkg = pkg
			if err = c.handler(f.getID(), buff); err != nil {
				f.error(fmt.Sprintf("%s handler err: %v", *clientID, err))
				return
			}

			pkgPool.Put(pkg)
			wrapPool.Put(buff)

			// clear mem
			w = nil
			err = nil
			buff = nil
			pkg = nil
		case <-c.ctx.Done():
			logs.Info(f.id, *clientID, " fwd reading loop was closed")
			return
		}
	}
}

func (f *Forwarder) handleKeyFrame(data []byte) {
	f.keyframeChann <- data
}

func (f *Forwarder) setKeyFrame(data *rtp.Packet) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.keyframe = data
}

// GetKeyFrame get key frame foreach fwd
func (f *Forwarder) GetKeyFrame() *rtp.Packet {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.keyframe
}

func (f *Forwarder) serveKeyFrame() {
	for {
		data, open := <-f.keyframeChann
		if !open {
			return
		}

		// get rtp pkg
		pkg := rtp.Packet{}
		err := pkg.Unmarshal(data)
		if err != nil {
			f.info(fmt.Sprintf("Unmarshal keyframe err: %v", err.Error()))
			continue
		}

		// parsing codec pkg
		switch f.codec {
		case MimeTypeVP8:
			err = f.handleVP8(&pkg)
		case MimeTypeVP9:
			err = f.handleVP9(&pkg)
		case MimeTypeH264:
			err = f.handleH264(&pkg)
		default:
			continue
		}

		if err != nil {
			f.info(fmt.Sprintf("Handle pkg data err: %v", err.Error()))
			continue
		}
	}
}

func (f *Forwarder) handleVP8(data *rtp.Packet) error {
	vp8Packet := &codecs.VP8Packet{}
	raw, err := vp8Packet.Unmarshal(data.Payload)
	if err != nil {
		return err
	}

	if IsVP8Keyframe(raw) {
		f.setKeyFrame(data)
		f.info(fmt.Sprintf("Save keyframe info (%s_%d_%d)", f.codec, data.PayloadType, data.SequenceNumber))
	}
	return nil
}

func (f *Forwarder) handleVP9(data *rtp.Packet) error {
	vp8Packet := &codecs.VP9Packet{}
	raw, err := vp8Packet.Unmarshal(data.Payload)
	if err != nil {
		return err
	}

	if IsVP9Keyframe(raw) {
		f.setKeyFrame(data)
		f.info(fmt.Sprintf("Save keyframe info (%s_%d_%d)", f.codec, data.PayloadType, data.SequenceNumber))
	}
	return nil
}

func (f *Forwarder) handleH264(data *rtp.Packet) error {
	vp8Packet := &codecs.H264Packet{}
	raw, err := vp8Packet.Unmarshal(data.Payload)
	if err != nil {
		return err
	}

	if IsH264Keyframe(raw) {
		f.setKeyFrame(data)
		f.info(fmt.Sprintf("Save keyframe info (%s_%d_%d)", f.codec, data.PayloadType, data.SequenceNumber))
	}
	return nil
}

// isKeyFrame := vp8Packet.Payload[0] & 0x01
// switch {
// case !i.seenKeyFrame && isKeyFrame == 1:
// 	return nil
// case i.currentFrame == nil && vp8Packet.S != 1:
// 	return nil
// }

// // Giả sử bạn có 'vp9Packet' đại diện cho gói tin VP9 bạn đã nhận được.
// vp9PayloadHeader := vp9Packet.Payload[0] // Trích xuất byte đầu tiên của dữ liệu VP9
// // Kiểm tra xem bit 'I' (Inter-picture predicted frame) có được đặt hay không
// isKeyFrame := (vp9PayloadHeader & 0x01) == 0

// if isKeyFrame {
//     // Đây là khung chính (I-frame)
//     // Thực hiện hành động thích hợp
// }
