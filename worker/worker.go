package worker

import (
	"fmt"
	"sync"

	"github.com/lamhai1401/gologs/loki"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/spgnk/rtc/errs"
	"github.com/spgnk/rtc/peer"
	"github.com/spgnk/rtc/utils"
)

// UpPeer to save mapping peer connection with video/audio up list
type UpPeer struct {
	audioTrackIDs map[string]string
	videoTrackIDs map[string]string
	mutex         sync.RWMutex
}

func (u *UpPeer) setAudioInList(trackID, codec *string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.audioTrackIDs[*trackID] = *codec
}

func (u *UpPeer) setVideoInList(trackID, codec *string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.videoTrackIDs[*trackID] = *codec
}

// GetVideoList linter
func (u *UpPeer) GetVideoList() map[string]string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.videoTrackIDs
}

// GetAudioList linter
func (u *UpPeer) GetAudioList() map[string]string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.audioTrackIDs
}

// SetVideoList linter
func (u *UpPeer) SetVideoList(arr map[string]string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.videoTrackIDs = arr
}

// SetAudioList linter
func (u *UpPeer) SetAudioList(arr map[string]string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.audioTrackIDs = arr
}

// GetVideoArr linter
func (u *UpPeer) GetVideoArr() []string {
	temp := make([]string, 0)
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	for trackID := range u.videoTrackIDs {
		temp = append(temp, trackID)
	}
	return temp
}

// GetAudioArr linter
func (u *UpPeer) GetAudioArr() []string {
	temp := make([]string, 0)
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	for trackID := range u.audioTrackIDs {
		temp = append(temp, trackID)
	}
	return temp
}

var _ (Worker) = (*PeerWorker)(nil)

// PeerWorker Set
type PeerWorker struct {
	id                  *string
	audioFwdm           utils.Fwdm                     // forward audio pkg
	videoFwdm           utils.Fwdm                     // forward video pkg
	peers               *utils.AdvanceMap              // save all peers with signalID
	upList              map[string]*UpPeer             // handler all stream obj
	tracks              map[string]*webrtc.TrackRemote // save trackID - obj to check has remote track or not
	handleNoConnection  func(signalID *string)
	trackMeta           map[string]bool // save track meta for detach
	readDeadlineHandler func(pcID, trackID *string, codec, kind string)
	mutex               sync.RWMutex
	logger              loki.Log
}

// NewPeerWorker linter
func NewPeerWorker(
	id *string,
	upList map[string]*UpPeer,
	logger loki.Log,
) Worker {
	w := &PeerWorker{
		id:        id,
		audioFwdm: utils.NewForwarderMannager(*id),
		videoFwdm: utils.NewForwarderMannager(*id),
		peers:     utils.NewAdvanceMap(),
		tracks:    make(map[string]*webrtc.TrackRemote),
		trackMeta: make(map[string]bool),
		upList:    upList,
		logger:    logger,
	}

	return w
}

// SetHandleNoConnection use for check peers doest have any peer
func (w *PeerWorker) SetHandleNoConnection(handler func(signalID *string)) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.handleNoConnection = handler
}

// Start linter
func (w *PeerWorker) Start() error {
	go w.countInterVal()
	return nil
}

// AddConnections add new connections
func (w *PeerWorker) AddConnections(signalID *string) {
	connections := peer.NewPeers(signalID, w.logger)
	if peers := w.getPeers(); peers != nil {
		peers.Set(*signalID, connections)
	}
}

// GetConnection couble be nil if not exist
func (w *PeerWorker) GetConnection(signalID, peerConnectionID *string) (*peer.Peer, error) {
	conns := w.getConnections(signalID)
	if conns == nil {
		return nil, fmt.Errorf("%s %s", *signalID, errs.ErrW001.Error())
	}

	return conns.GetConnection(peerConnectionID), nil
}

// GetConnections could be nil if not exist
func (w *PeerWorker) GetConnections(signalID *string) peer.Connections {
	return w.getConnections(signalID)
}

// RemoveConnections remove all connections
func (w *PeerWorker) RemoveConnections(signalID *string) {
	w.closeConnections(signalID)
}

// RemoveConnection remove existing peer connection
func (w *PeerWorker) RemoveConnection(signalID, peerConnectionID, cookieID *string) error {
	// get connections
	connections := w.getConnections(signalID)
	if connections == nil {
		return fmt.Errorf("%s %s", *signalID, errs.ErrW001.Error())
	}
	conn := connections.GetConnection(peerConnectionID)
	if conn != nil {
		if *conn.GetCookieID() == *cookieID {
			connections.RemoveConnection(peerConnectionID)
		} else {
			conn.Error(fmt.Sprintf("input cookieID != peer cookieID (%s_%s). Dont remove", *cookieID, *conn.GetCookieID()), map[string]any{
				"signal_id": *signalID,
			})
		}
	}

	return nil
}

// AddDCConnection add new peer connection
func (w *PeerWorker) AddDCConnection(
	signalID *string,
	configs *peer.Configs,
	handleOnDatachannel func(pcID *string, d *webrtc.DataChannel),
	handleAddDCPeer func(signalID, role, peerConnectionID *string),
	handleFailedDCPeer func(signalID, role, peerConnectionID *string),
	handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
) (*peer.Peer, error) {
	// get connections
	connections := w.getConnections(signalID)
	if connections == nil {
		return nil, fmt.Errorf("%s %s", *signalID, errs.ErrW001.Error())
	}
	conn, err := connections.AddDCConnection(
		configs,
		handleOnDatachannel,
		handleAddDCPeer,
		handleFailedDCPeer,
		handleCandidate,
	)

	if err != nil {
		return nil, fmt.Errorf("add new dc connection %s err: %s", *configs.PeerConnectionID, err.Error())
	}
	return conn, nil
}

// AddConnection add new peer connection
func (w *PeerWorker) AddConnection(
	signalID *string,
	configs *peer.Configs,
	handleAddPeer func(signalID, role, peerConnectionID *string),
	handleFailedPeer func(signalID, role, peerConnectionID *string),
	handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
	handleOnNegotiationNeeded func(signalID, peerConnectionID, cookieID *string),
) (*peer.Peer, error) {
	// get connections
	connections := w.getConnections(signalID)
	if connections == nil {
		return nil, fmt.Errorf("%s %s", *signalID, errs.ErrW001.Error())
	}

	conn, err := connections.AddConnection(
		configs,
		w.handleOnTrack,
		handleAddPeer,
		handleFailedPeer,
		handleCandidate,
		handleOnNegotiationNeeded,
	)

	if err != nil {
		return nil, fmt.Errorf("add new connection %s err: %s", *configs.PeerConnectionID, err.Error())
	}
	return conn, nil
}

// AddVideoFwd to add new video fwd
func (w *PeerWorker) AddVideoFwd(trackID *string) {
	// if fwdm := w.getVideoFwdm(); fwdm != nil {
	// 	fwdm.AddNewForwarder(*trackID)
	// }
	w.videoFwdm.AddNewForwarder(*trackID)
}

// AddAudioFwd to add new video fwd
func (w *PeerWorker) AddAudioFwd(trackID *string) {
	// if fwdm := w.getAudioFwdm(); fwdm != nil {
	// 	fwdm.AddNewForwarder(*trackID)
	// }
	w.audioFwdm.AddNewForwarder(*trackID)
}

// UnRegister all video and audio
func (w *PeerWorker) UnRegister(peerConnectionID *string) {
	// if fwdm := w.getVideoFwdm(); fwdm != nil {
	// 	fwdm.UnregisterAll(*peerConnectionID)
	// 	w.Warn(fmt.Sprintf("%s unRegister all VideoFwdm", *peerConnectionID))
	// }

	w.videoFwdm.UnregisterAll(*peerConnectionID)
	w.Warn(fmt.Sprintf("%s unRegister all VideoFwdm", *peerConnectionID))

	w.audioFwdm.UnregisterAll(*peerConnectionID)
	w.Warn(fmt.Sprintf("%s unRegister all AudioFwdm", *peerConnectionID))
	// if fwdm := w.getAudioFwdm(); fwdm != nil {
	// 	fwdm.UnregisterAll(*peerConnectionID)
	// 	w.Warn(fmt.Sprintf("%s unRegister all AudioFwdm", *peerConnectionID))
	// }
}

// UnRegisterVideo linter
func (w *PeerWorker) UnRegisterVideo(peerConnectionID, videoTrackID *string) {
	if videoTrackID != nil && *videoTrackID != "" {
		// check duplicate register
		// fwdm := w.getVideoFwdm()
		// if fwdm != nil {
		// 	fwdm.Unregister(videoTrackID, peerConnectionID)
		// 	w.Warn(fmt.Sprintf("%s unRegister (%s) of VideoFwdm", *peerConnectionID, *videoTrackID))
		// }

		w.videoFwdm.Unregister(videoTrackID, peerConnectionID)
		w.Warn(fmt.Sprintf("%s unRegister (%s) of VideoFwdm", *peerConnectionID, *videoTrackID))
	}

}

// UnRegisterAudio linter
func (w *PeerWorker) UnRegisterAudio(peerConnectionID, audioTrackID *string) {
	if audioTrackID != nil && *audioTrackID != "" {
		// fwdm := w.getAudioFwdm()
		// if fwdm != nil {
		// 	fwdm.Unregister(audioTrackID, peerConnectionID)
		// 	w.Warn(fmt.Sprintf("%s unRegister (%s) in AudioFwdm", *peerConnectionID, *audioTrackID))
		// }
		w.audioFwdm.Unregister(audioTrackID, peerConnectionID)
		w.Warn(fmt.Sprintf("%s unRegister (%s) in AudioFwdm", *peerConnectionID, *audioTrackID))
	}
}

// Register a client to all video/audio fwd
func (w *PeerWorker) Register(
	signalID,
	peerConnectionID *string,
	videoTrackIDs,
	audioTrackIDs []string,
	errHandler func(signalID, peerConnectionID, trackID *string, reason string),
) error {
	p := w.getPeer(signalID, peerConnectionID)
	if p == nil {
		return fmt.Errorf("[%s-%s] %s", *signalID, *peerConnectionID, errs.ErrP002)
	}

	// w.videoFwdm.UnregisterAll(*peerConnectionID)
	for _, v := range videoTrackIDs {
		err := w.RegisterVideo(signalID, &v, peerConnectionID, errHandler)
		if err != nil {
			return err
		}
	}

	// w.audioFwdm.UnregisterAll(*peerConnectionID)
	for _, v := range audioTrackIDs {
		err := w.RegisterAudio(signalID, &v, peerConnectionID, errHandler)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterVideo linter
func (w *PeerWorker) RegisterVideo(
	signalID,
	videoTrackID,
	peerConnectionID *string,
	errHandler func(signalID, peerConnectionID, trackID *string, reason string),
) error {
	p := w.getPeer(signalID, peerConnectionID)
	if p == nil {
		return fmt.Errorf("[%s-%s] %s", *signalID, *peerConnectionID, errs.ErrP002)
	}

	// fwdm := w.getVideoFwdm()
	// if fwdm == nil {
	// 	return fmt.Errorf("video fwdm is nil")
	// }

	// check duplicate register
	// isDupLicate := fwdm.GetClient(videoTrackID, peerConnectionID)
	// if isDupLicate != nil {
	// 	return fmt.Errorf("%s_%s register duplicated", *peerConnectionID, *videoTrackID)
	// }

	videoHandler := func(trackID string, wrapper *utils.Wrapper) error {
		// if !p.IsConnected() {
		// 	return nil
		// }

		// if trackID != "t-1--v" {
		// 	return nil
		// }

		err := p.AddVideoRTP(&trackID, peerConnectionID, wrapper.Pkg)
		if err != nil {
			errHandler(signalID, peerConnectionID, &trackID, err.Error())
			return err
		}

		wrapper = nil
		return nil
	}

	// w.videoFwdm.Unregister(videoTrackID, p.GetPeerConnectionID())
	w.videoFwdm.Register(*videoTrackID, *peerConnectionID, videoHandler)
	return nil
}

// RegisterAudio linter
func (w *PeerWorker) RegisterAudio(
	signalID,
	audioTrackID,
	peerConnectionID *string,
	errHandler func(signalID, peerConnectionID, trackID *string, reason string),
) error {
	p := w.getPeer(signalID, peerConnectionID)
	if p == nil {
		return fmt.Errorf("[%s-%s] %s", *signalID, *peerConnectionID, errs.ErrP002)
	}

	// fwdm := w.getAudioFwdm()
	// if fwdm == nil {
	// 	return fmt.Errorf("audio fwdm is nil")
	// }

	// check duplicate register
	// isDupLicate := fwdm.GetClient(audioTrackID, peerConnectionID)
	// if isDupLicate != nil {
	// 	return fmt.Errorf("%s_%s register duplicated", *peerConnectionID, *audioTrackID)
	// }

	audioHandler := func(trackID string, wrapper *utils.Wrapper) error {
		// if !p.IsConnected() {
		// 	return nil
		// }

		// if trackID != "t-1--a" {
		// 	return nil
		// }

		err := p.AddAudioRTP(&trackID, peerConnectionID, wrapper.Pkg)
		if err != nil {
			errHandler(signalID, peerConnectionID, &trackID, err.Error())
			return err
		}
		// w.Stack(fmt.Sprintf("Write %s audio rtp to %s", trackID, peerConnectionID))
		wrapper = nil
		return nil
	}

	// w.audioFwdm.Unregister(audioTrackID, p.GetPeerConnectionID())
	w.audioFwdm.Register(*audioTrackID, *peerConnectionID, audioHandler)
	return nil
}

// GetStates return all pcID - states
func (w *PeerWorker) GetStates() map[string]string {
	temp := make(map[string]string)
	if peers := w.getPeers(); peers != nil {
		peers.Iter(func(_, value interface{}) bool {
			connections, ok := value.(peer.Connections)
			if ok {
				connections.GetStates(temp)
			}

			return true
		})
	}

	return temp
}

// GetAllConnectionID return signalID - [streamID]
func (w *PeerWorker) GetAllConnectionID() map[string][]string {
	tmp := make(map[string][]string)
	if ps := w.getPeers(); ps != nil {
		ps.Iter(func(key, value interface{}) bool {
			signalID, ok1 := key.(string)
			connections, ok2 := value.(peer.Connections)
			if ok1 && ok2 {
				tmp[signalID] = connections.GetAllPeerConnectionID()
			}
			return true
		})
	}
	return tmp
}

// SetUpList linter
func (w *PeerWorker) SetUpList(lst map[string]*UpPeer) {
	w.setUpList(lst)
}

// DeleteUpList delete peerUp
func (w *PeerWorker) DeleteUpList(peerConnectionID *string) {
	w.deleteUpList(peerConnectionID)
}

// AddUpList add new PeerUP
func (w *PeerWorker) AddUpList(peerConnectionID *string, c *UpPeer) {
	w.addUpList(peerConnectionID, c)
}

// AppendUpList linter
func (w *PeerWorker) AppendUpList(pcID *string, obj *UpPeer) {
	currObj := w.getUpPeer(pcID)

	if currObj == nil {
		w.setUpPeer(pcID, obj)
		return
	}

	for trackID, codec := range obj.audioTrackIDs {
		currObj.setAudioInList(&trackID, &codec)
	}

	for trackID, codec := range obj.videoTrackIDs {
		currObj.setVideoInList(&trackID, &codec)
	}
}

// handle peer remotetrack with streamID
func (w *PeerWorker) handleOnTrack(signalID, peerConnectionID *string, remoteTrack *webrtc.TrackRemote) {
	kind := remoteTrack.Kind().String()
	// find trackId in stream ọbject
	trackID, err := w.findTrackID(peerConnectionID, &kind)
	codec := remoteTrack.Codec().MimeType
	if err != nil {
		w.Warn(err.Error() + " so use remoteTrackID")
		trackID = remoteTrack.ID()
	}

	w.Info(fmt.Sprintf("(%s_%s) Has remote track of id %s_%s", trackID, codec, *signalID, *peerConnectionID))

	var fwdm utils.Fwdm
	switch kind {
	case "video":
		fwdm = w.videoFwdm
	case "audio":
		fwdm = w.audioFwdm
	default:
		return
	}

	go w.pushToFwd(fwdm, remoteTrack, &trackID, &kind, peerConnectionID)
}

// ReadRTP is a convenience method that wraps Read and unmarshals for you.
// func (t *TrackRemote) ReadRTP() (*rtp.Packet, interceptor.Attributes, error) {
// 	b := make([]byte, t.receiver.api.settingEngine.getReceiveMTU())
// 	i, attributes, err := t.Read(b)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	r := &rtp.Packet{}
// 	if err := r.Unmarshal(b[:i]); err != nil {
// 		return nil, nil, err
// 	}
// 	return r, attributes, nil
// }

func (w *PeerWorker) pushToFwd(fwdm utils.Fwdm, remoteTrack *webrtc.TrackRemote, trackID, kind, peerConnectionID *string) {
	var pkg *rtp.Packet
	var err error
	var i int
	var b *[]byte
	codec := remoteTrack.Codec().MimeType

	// readDeadLine := 16 * time.Second

	defer func() {
		w.deleteTrackMeta(*trackID)
	}()

	var rlBufPool = sync.Pool{New: func() interface{} {
		buf := make([]byte, 1460)
		return &buf
	}}

	var pkgPool = sync.Pool{New: func() interface{} {
		return &rtp.Packet{}
	}}

	w.setRemoteTrack(trackID, remoteTrack)
	defer w.deleteRemoteTrack(trackID)
	for {
		b = rlBufPool.Get().(*[]byte)
		// err = remoteTrack.SetReadDeadline(time.Now().Add(readDeadLine))
		// if err != nil {
		// 	w.Error("SetReadDeadline err: ", err.Error())
		// }
		i, _, err = remoteTrack.Read(*b)
		if err != nil {
			// if err == os.ErrDeadlineExceeded || strings.Contains(err.Error(), os.ErrDeadlineExceeded.Error()) {
			// 	// signal to worker
			// 	w.Warn(fmt.Sprintf("%s_%s readDeadline exceeded.", *peerConnectionID, *trackID))
			// 	w.readDeadlineHandler(peerConnectionID, trackID, codec, *kind)
			// }
			return
		}

		// get rtp pkb
		// pkg = pkgPool.Get().(*rtp.Packet)
		// if err = pkg.Unmarshal(b[:i]); err != nil {
		// 	w.Error(err.Error())
		// 	return
		// }

		// pkg, _, err = remoteTrack.ReadRTP()
		// if err != nil {
		// 	w.Error(fmt.Sprintf("%s readRTP error %s", *peerConnectionID, err.Error()))
		// 	return
		// }

		// push video to fwd
		fwd := fwdm.GetForwarder(*trackID)
		if fwd == nil {
			fwd = fwdm.AddNewForwarder(*trackID)
		}

		// pushing data to fwd
		if fwd != nil {
			fwd.Push(&utils.Wrapper{
				// Pkg: pkg,
				Data: (*b)[:i],
			})
			w.Stack(fmt.Sprintf("%s_%s Push rtp pkg to fwd %s", *peerConnectionID, codec, *trackID))
		}

		*b = make([]byte, 1460)
		// push data back
		rlBufPool.Put(b)
		pkgPool.Put(pkg)

		pkg = nil
		err = nil
		fwd = nil
		b = nil
	}
}

// findTrackID use only for peer up
func (w *PeerWorker) findTrackID(peerConnectionID, kind *string) (string, error) {
	var id string
	lst := w.getUpList()
	if lst == nil {
		return id, errs.ErrPS0041
	}

	obj := lst[*peerConnectionID]
	if obj == nil {
		return id, fmt.Errorf("%s %s", *peerConnectionID, errs.ErrPS0042.Error())
	}

	var arr []string
	switch *kind {
	case "video":
		arr = obj.GetVideoArr()
	case "audio":
		arr = obj.GetAudioArr()
	default:
		return id, fmt.Errorf("wrong kind track id: %s", *kind)
	}

	length := len(arr)
	switch length {
	case 0:
		return id, fmt.Errorf("%s %s %s", *peerConnectionID, *kind, errs.ErrPS0043.Error())
	case 1:
		id = arr[0] // there is only one id in arr if peer up
	default:
		return id, fmt.Errorf("cannot find track id in %s track ids with current length %d", *kind, length)
	}

	return id, nil
}

// GetRemoteTrack linter
func (w *PeerWorker) GetRemoteTrack(trackID *string) *webrtc.TrackRemote {
	return w.getRemoteTrack(trackID)
}

// GetVideoReceiveTime linter
func (w *PeerWorker) GetVideoReceiveTime() map[string]int64 {
	return w.videoFwdm.GetLastTimeReceive()
}

// GetVideoReceiveTimeby linter
func (w *PeerWorker) GetVideoReceiveTimeby(trackID string) int64 {
	return w.videoFwdm.GetLastTimeReceiveBy(trackID)
}

// GetAudioReceiveTime linter
func (w *PeerWorker) GetAudioReceiveTime() map[string]int64 {
	return w.audioFwdm.GetLastTimeReceive()
}

// GetAudioReceiveTimeby linter
func (w *PeerWorker) GetAudioReceiveTimeby(trackID string) int64 {
	return w.audioFwdm.GetLastTimeReceiveBy(trackID)
}

// SetHandleReadDeadline linter
func (w *PeerWorker) SetHandleReadDeadline(f func(pcID, trackID *string, codec, kind string)) {
	w.mutex.Lock()
	w.readDeadlineHandler = f
	w.mutex.Unlock()
}

func (w *PeerWorker) handleReadDeadlinefunc(pcID, trackID *string, codec, kind string) {
	if state := w.GetTrackMeta(*trackID); state {
		go w.readDeadlineHandler(pcID, trackID, codec, kind)
	}
}

func (w *PeerWorker) deleteTrackMeta(trackID string) {
	w.mutex.Lock()
	delete(w.trackMeta, trackID)
	w.mutex.Unlock()
}

// GetTrackMeta to check track meta
func (w *PeerWorker) GetTrackMeta(trackID string) bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.trackMeta[trackID]

}

// SetTrackMeta to set track meta
func (w *PeerWorker) SetTrackMeta(trackID string, state bool) {
	w.mutex.Lock()
	w.trackMeta[trackID] = state
	w.mutex.Unlock()
}
