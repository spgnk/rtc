package peer

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/lamhai1401/gologs/loki"
	"github.com/spgnk/rtc/errs"
	"github.com/spgnk/rtc/utils"

	"github.com/mitchellh/mapstructure"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

// Peer linter
type Peer struct {
	cookieID    *string           // internal cookieID to check remove
	bitrate     *int              // bitrate
	iceCache    *utils.AdvanceMap // save all ice before set remote description
	isConnected bool              // check to this is connection init or ice restart
	isClosed    bool              // check peer close or not

	// peer mission for allow up/down video data
	config *Configs

	tracks *LocalTracks // handle all create/remove localTrack

	debug string

	remoteTrack *webrtc.TrackRemote // save last remoteTrack

	conn  *webrtc.PeerConnection
	mutex sync.RWMutex

	duplicated map[string]bool
	pli        int // set PLI interval

	logger loki.Log
}

// NewPeerConnection linter
func NewPeerConnection(configs *Configs) Connection {
	return newPeerConnection(configs)
}

func newPeerConnection(configs *Configs) *Peer {
	cookieID := utils.GenerateID()
	p := &Peer{
		cookieID:    &cookieID,
		isConnected: false,
		isClosed:    false,
		iceCache:    utils.NewAdvanceMap(),
		config:      configs,
		debug:       os.Getenv("DEBUG"),
		duplicated:  make(map[string]bool),
		logger:      configs.logger,
	}

	if configs.Bitrate == nil {
		br := 200
		p.bitrate = &br
	}
	return p
}

// SetCodecPreferences linter
func (p *Peer) SetCodecPreferences(payLoadType *int, trackConfig *TrackConfig) error {
	return p.tracks.setCodecPreferences(payLoadType, trackConfig)
}

// InitDC linter
func (p *Peer) InitDC(
	handleOnDatachannel func(pcID *string, d *webrtc.DataChannel),
) (*webrtc.PeerConnection, error) {

	// s := webrtc.SettingEngine{}
	// s.DetachDataChannels()

	// Create an API object with the engine
	// api := webrtc.NewAPI(webrtc.WithSettingEngine(s))

	// Create a new RTCPeerConnection using the API object
	// peerConnection, err := api.NewPeerConnection(*p.config.TurnConfig)

	// Create a new PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(*p.config.TurnConfig)
	if err != nil {
		return nil, err
	}

	if !p.config.IsCreateDC {
		// Register data channel creation handling
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
			p.Info(fmt.Sprintf("New DataChannel %s %d\n", d.Label(), d.ID()))
			p.Warn("receive DC, handle open")

			// Register channel opening handling
			d.OnOpen(func() {
				p.Warn(fmt.Sprintf("Data channel '%s'-'%d' open.\n", d.Label(), d.ID()))
				handleOnDatachannel(p.GetPeerConnectionID(), d)

				// Detach the data channel
				// raw, dErr := d.Detach()
				// if dErr != nil {
				// 	panic(dErr)
				// }

				// // Handle reading from the data channel
				// if readLoop != nil {
				// 	go readLoop(p.config.PeerConnectionID, raw)
				// }

				// // Handle writing to the data channel
				// if writeLoop != nil {
				// 	go writeLoop(p.config.PeerConnectionID, raw)
				// }
			})
		})
	} else {

		// maxRetransmits := uint16(0)
		// options := &webrtc.DataChannelInit{
		// Ordered:        &ordered,
		// MaxRetransmits: &maxRetransmits,
		// }
		dataChannel, err := peerConnection.CreateDataChannel("haideptrai", nil)
		if err != nil {
			return nil, err
		}

		// Register channel opening handling
		dataChannel.OnOpen(func() {
			p.Warn(fmt.Sprintf("Data channel '%s'-'%d' open.\n", dataChannel.Label(), dataChannel.ID()))
			handleOnDatachannel(p.GetPeerConnectionID(), dataChannel)

			// Detach the data channel
			// raw, dErr := dataChannel.Detach()
			// if dErr != nil {
			// 	logs.Error(err.Error())
			// }

			// // Handle reading from the data channel
			// if readLoop != nil {
			// 	go readLoop(p.config.PeerConnectionID, raw)
			// }

			// // Handle writing to the data channel
			// if writeLoop != nil {
			// 	go writeLoop(p.config.PeerConnectionID, raw)
			// }
		})
	}

	p.setConn(peerConnection)
	return peerConnection, nil
}

// Init linter
func (p *Peer) Init() (*webrtc.PeerConnection, error) {
	api, err := p.initAPI(p.config)
	if err != nil {
		return nil, err
	}
	// init peer
	conn, err := api.NewPeerConnection(*p.config.TurnConfig)
	if err != nil {
		return nil, err
	}
	p.setConn(conn)

	tracks := NewTracks(conn)
	p.tracks = tracks

	return conn, nil
}

// Close close peer connection
func (p *Peer) Close() {
	if !p.checkClose() {
		p.setClose(true)
		p.setBitrate(nil)
		if err := p.closeConn(); err != nil {
			p.Error(err.Error())
		}
		// logs.Warn(fmt.Sprintf("%s_%s conn was closed", p.getPeerConnectionID(), p.getCookieID()))
	}
}

// SetIsConnected linter
func (p *Peer) SetIsConnected(states bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isConnected = states
}

// IsConnected check this is init connection or retrieve
func (p *Peer) IsConnected() bool {
	return p.isConnected
}

// GetRole linter
func (p *Peer) GetRole() *string {
	return p.getRole()
}

// GetCookieID linter
func (p *Peer) GetCookieID() *string {
	return p.getCookieID()
}

// GetPeerConnectionID linter
func (p *Peer) GetPeerConnectionID() *string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.config.PeerConnectionID
}

// GetLocalDescription get current peer local description to send to client
func (p *Peer) GetLocalDescription() (*webrtc.SessionDescription, error) {
	return p.getLocalDescription()
}

// AddSDP add sdp, input raw data or utils.SDPTemp
func (p *Peer) AddSDP(values interface{}) error {

	conns := p.getConn()
	if conns == nil {
		return errs.ErrP002
	}

	data, ok := values.(utils.SDPTemp)
	if !ok {
		err := mapstructure.Decode(values, &data)
		if err != nil {
			return err
		}
	}

	sdp := &webrtc.SessionDescription{
		Type: utils.NewSDPType(data.Type),
		SDP:  data.SDP,
	}

	switch data.Type {
	case "offer":
		if err := p.AddOffer(sdp); err != nil {
			return err
		}
	case "answer":
		if err := p.AddAnswer(sdp); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid sdp type: %s", data.Type)
	}

	return nil
}

// AddOffer add client offer and return answer
func (p *Peer) AddOffer(offer *webrtc.SessionDescription) error {
	conn := p.getConn()
	if conn == nil {
		return errs.ErrP002
	}

	// set remote desc
	err := conn.SetRemoteDescription(*offer)
	if err != nil {
		return err
	}

	err = p.setCacheIce()
	if err != nil {
		return err
	}

	err = p.CreateAnswer()
	if err != nil {
		return err
	}

	return nil
}

// AddAnswer add client answer and set remote desc
func (p *Peer) AddAnswer(answer *webrtc.SessionDescription) error {
	conn := p.getConn()
	if conn == nil {
		return errs.ErrP002
	}

	// set remote desc
	err := conn.SetRemoteDescription(*answer)
	if err != nil {
		return err
	}
	return p.setCacheIce()
}

// AddICECandidate to add candidate
func (p *Peer) AddICECandidate(icecandidate interface{}) error {
	// var candidateInit webrtc.ICECandidateInit
	candidateInit, ok := icecandidate.(*webrtc.ICECandidateInit)
	if !ok {
		err := mapstructure.Decode(icecandidate, &candidateInit)
		if err != nil {
			return err
		}
	}

	conn := p.getConn()
	if conn == nil {
		p.addIceCache(candidateInit)
		return errs.ErrP002
	}

	if conn.RemoteDescription() == nil {
		p.addIceCache(candidateInit)
		p.Info(fmt.Sprintf("%s Add candidate to cache because remote is nil or peer states is %v", *p.GetPeerConnectionID(), p.IsConnected()))
		return nil
	}

	err := conn.AddICECandidate(*candidateInit)
	if err != nil {
		return err
	}

	return nil
}

// AddVideoRTP linter
func (p *Peer) AddVideoRTP(trackID, peerConnectionID *string, packet *rtp.Packet) error {
	// parse to source or dest to get local video
	track := p.tracks.getVideoTrack(trackID)
	if track == nil {
		return errs.ErrP0031
	}

	if err := p.writeRTP(packet, track); err != nil {
		return err
	}

	if !p.tracks.getReceiveData(trackID) {
		p.tracks.setReceiveData(trackID, true)
	}

	if p.tracks.getFirstInitTrack(trackID) == "1" {
		p.tracks.setFirstInitTrack(trackID, "2")
	}
	p.stackDebug(fmt.Sprintf("Write video rtp to codec/track/pcID (%s_%s_%s)", track.Codec().MimeType, *trackID, *p.GetPeerConnectionID()))
	return nil
}

// AddAudioRTP linter
func (p *Peer) AddAudioRTP(trackID, peerConnectionID *string, packet *rtp.Packet) error {
	track := p.tracks.getAudioTrack(trackID)
	if track == nil {
		return errs.ErrP0032
	}
	if err := p.writeRTP(packet, track); err != nil {
		return err
	}

	if !p.tracks.getReceiveData(trackID) {
		p.tracks.setReceiveData(trackID, true)
	}

	if p.tracks.getFirstInitTrack(trackID) == "1" {
		p.tracks.setFirstInitTrack(trackID, "2")
	}
	p.stackDebug(fmt.Sprintf("Write audio rtp to codec/track/pcID (%s_%s_%s)", track.Codec().MimeType, *trackID, *p.GetPeerConnectionID()))
	return nil
}

// getNodeLevel get level of current node, default is 0
func getNodeLevel() int {
	nodeLevel := os.Getenv("NODE_LEVEL")
	if nodeLevel == "" {
		return 0
	}

	level, err := strconv.Atoi(nodeLevel)
	if err != nil {
		return 0
	}

	return level
}

// HandleVideoTrack handle all video track
func (p *Peer) HandleVideoTrack(remoteTrack *webrtc.TrackRemote) {
	// go p.modifyBitrate(remoteTrack)
	if getNodeLevel() == 0 {
		go p.pictureLossIndication(remoteTrack)
	}

	// setRemoteTrack to set PLI on the mand
	p.setRemoteTrack(remoteTrack)
}

// CreateOffer add offer
func (p *Peer) CreateOffer(iceRestart bool) error {
	conn := p.getConn()
	if conn == nil {
		return errs.ErrP002
	}

	var opt *webrtc.OfferOptions
	if iceRestart {
		opt = &webrtc.OfferOptions{
			ICERestart: iceRestart,
		}
	}
	// set local desc
	offer, err := conn.CreateOffer(opt)
	if err != nil {
		return err
	}

	err = conn.SetLocalDescription(offer)
	if err != nil {
		return err
	}
	return nil
}

// CreateAnswer add answer
func (p *Peer) CreateAnswer() error {
	conn := p.getConn()
	if conn == nil {
		return errs.ErrP002
	}
	// set local desc
	answer, err := conn.CreateAnswer(nil)
	if err != nil {
		return err
	}

	err = conn.SetLocalDescription(answer)
	if err != nil {
		return err
	}
	return nil
}

// AddVideoTrack linter
func (p *Peer) AddVideoTrack(trackConfig *TrackConfig) error {
	return p.tracks.createLocalVideo(trackConfig)
}

// AddAudioTrack linter
func (p *Peer) AddAudioTrack(trackConfig *TrackConfig) error {
	return p.tracks.createLocalAudio(trackConfig)
}

// RemoveVideoTrack remove peering existing track
func (p *Peer) RemoveVideoTrack(trackID *string) error {
	var err error
	conn := p.getConn()
	if conn == nil {
		return errs.ErrP002
	}

	// get sender from tracks
	sender := p.tracks.getVideoSender(trackID)
	if sender == nil {
		return fmt.Errorf("RemoveVideoTrack %s ", errs.ErrP005.Error())
	}

	err = conn.RemoveTrack(sender)
	if err != nil {
		return err
	}
	err = p.tracks.removeLocalVideoTrack(trackID)
	if err != nil {
		return err
	}

	return nil
}

// GetAudioRTPTrack linter
func (p *Peer) GetAudioRTPTrack(trackID *string) *webrtc.TrackLocalStaticRTP {
	return p.tracks.getAudioTrack(trackID)
}

// GetVideoRTPTrack linter
func (p *Peer) GetVideoRTPTrack(trackID *string) *webrtc.TrackLocalStaticRTP {
	return p.tracks.getVideoTrack(trackID)
}

// ReplaceAudioTrack remove peering existing track
func (p *Peer) ReplaceAudioTrack(trackID, codec *string) error {
	return p.tracks.replaceAudioRTPTrack(trackID, codec)
}

// ReplaceVideoTrack remove peering existing track
func (p *Peer) ReplaceVideoTrack(trackID, codec *string) error {
	return p.tracks.replaceVideoRTPTrack(trackID, codec)
}

// RemoveAudioTrack remove peering existing track
func (p *Peer) RemoveAudioTrack(trackID *string) error {
	var err error
	conn := p.getConn()
	if conn == nil {
		return errs.ErrP002
	}

	// get sender from tracks
	sender := p.tracks.getAudioSender(trackID)
	if sender == nil {
		return fmt.Errorf("RemoveAudioTrack %s ", errs.ErrP005.Error())
	}

	err = conn.RemoveTrack(sender)
	if err != nil {
		return err
	}
	err = p.tracks.removeLocalAudioTrack(trackID)
	if err != nil {
		return err
	}

	return nil
}

// AddVideoSample linter
func (p *Peer) AddVideoSample(trackID, peerConnectionID *string, sample *media.Sample) error {
	// parse to source or dest to get local video
	track := p.tracks.getVideoTrackSample(trackID)
	if track == nil {
		return errs.ErrP0033
	}

	if err := p.writeSample(sample, track); err != nil {
		return err
	}
	p.stackDebug(fmt.Sprintf("Write video sample to codec/track/pcID (%s_%s_%s)", track.Codec().MimeType, *trackID, *p.GetPeerConnectionID()))
	return nil
}

// AddAudioSample linter
func (p *Peer) AddAudioSample(trackID, peerConnectionID *string, sample *media.Sample) error {
	// parse to source or dest to get local video
	track := p.tracks.getAudioTrackSample(trackID)
	if track == nil {
		return errs.ErrP0034
	}

	if err := p.writeSample(sample, track); err != nil {
		return err
	}

	p.stackDebug(fmt.Sprintf("Write audio sample to codec/track/pcID (%s_%s_%s)", track.Codec().MimeType, *trackID, *p.GetPeerConnectionID()))
	return nil
}

// IsReceivedData linter
func (p *Peer) IsReceivedData(trackID *string) bool {
	return p.tracks.getReceiveData(trackID)
}

// IsFirstInit return 1 if is first init, 2 is received data and not
func (p *Peer) IsFirstInit(trackID *string) string {
	return p.tracks.getFirstInitTrack(trackID)
}

// SetPliInterval linter
func (p *Peer) SetPliInterval(interval int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.pli = interval
}
