package peer

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

// Connection linter
type Connection interface {
	Init() (*webrtc.PeerConnection, error)

	GetRole() *string
	GetCookieID() *string
	GetPeerConnectionID() *string
	GetLocalDescription() (*webrtc.SessionDescription, error)

	GetAudioRTPTrack(trackID *string) *webrtc.TrackLocalStaticRTP
	GetVideoRTPTrack(trackID *string) *webrtc.TrackLocalStaticRTP

	ReplaceAudioTrack(trackID, codec *string) error
	ReplaceVideoTrack(trackID, codec *string) error

	CreateOffer(iceRestart bool) error
	CreateAnswer() error
	Close()

	AddVideoTrack(trackConfig *TrackConfig) error
	AddAudioTrack(trackConfig *TrackConfig) error
	RemoveVideoTrack(trackID *string) error
	RemoveAudioTrack(trackID *string) error

	AddSDP(values interface{}) error
	AddOffer(offer *webrtc.SessionDescription) error
	AddAnswer(answer *webrtc.SessionDescription) error
	AddICECandidate(icecandidate interface{}) error

	AddVideoRTP(trackID, peerConnectionID *string, packet *rtp.Packet) error
	AddAudioRTP(trackID, peerConnectionID *string, packet *rtp.Packet) error
	AddVideoSample(trackID, peerConnectionID *string, sample *media.Sample) error
	AddAudioSample(trackID, peerConnectionID *string, sample *media.Sample) error

	HandleVideoTrack(remoteTrack *webrtc.TrackRemote)

	// SetCodecPreferences sets preferred list of supported codecs
	// if codecs is empty or nil we reset to default from MediaEngine
	SetCodecPreferences(payLoadType *int, trackConfig *TrackConfig) error

	IsConnected() bool
	SetIsConnected(states bool)

	// IsReceivedData linter
	IsReceivedData(trackID *string) bool
	// IsFirstInit return 1 if is first init, 2 is received data and not
	IsFirstInit(trackID *string) string

	InitDC(
		handleOnDatachannel func(pcID *string, d *webrtc.DataChannel),
	) (*webrtc.PeerConnection, error)

	SendPictureLossIndication()

	AddDuplicated(t string, element bool)
	GetDuplicated(t string) bool
	DeleteDuplicated(t string)

	// SetPliInterval linter
	SetPliInterval(int)
}

// Connections linter
type Connections interface {
	Close()
	AddConnection(
		configs *Configs,
		handleOnTrack func(signalID, peerConnectionID *string, remoteTrack *webrtc.TrackRemote),
		handleAddPeer func(signalID, role, peerConnectionID *string),
		handleFailedPeer func(signalID, role, peerConnectionID *string),
		handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
		handleOnNegotiationNeeded func(signalID, peerConnectionID, cookieID *string),
	) (*Peer, error)

	RemoveConnection(
		peerConnectionID *string,
	)

	GetAllConnection() []*Peer
	GetAllPeerConnectionID() []string
	GetConnection(peerConnectionID *string) *Peer
	GetStates(temp map[string]string)        // get all connection states
	GetState(peerConnectionID string) string // get single connection state

	// AddCandidate(peerConnectionID *string, value interface{}) error
	AddSDP(peerConnectionID *string, value interface{}) error

	CountAllPeer() int64

	AddDCConnection(
		configs *Configs,
		handleOnDatachannel func(pcID *string, d *webrtc.DataChannel),
		handleAddDCPeer func(signalID, role, peerConnectionID *string),
		handleFailedDCPeer func(signalID, role, peerConnectionID *string),
		handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
	) (*Peer, error)
}
