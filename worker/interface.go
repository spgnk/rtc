package worker

import (
	"github.com/pion/webrtc/v3"
	"github.com/spgnk/rtc/peer"
)

// Worker peer connection worker
type Worker interface {
	Start() error

	AddDCConnection(
		signalID *string,
		configs *peer.Configs,
		handleOnDatachannel func(pcID *string, d *webrtc.DataChannel),
		handleAddDCPeer func(signalID, role, peerConnectionID *string),
		handleFailedDCPeer func(signalID, role, peerConnectionID *string),
		handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
	) (*peer.Peer, error)

	AddConnection(
		signalID *string,
		configs *peer.Configs,
		handleAddPeer func(signalID, role, peerConnectionID *string),
		handleFailedPeer func(signalID, role, peerConnectionID *string),
		handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
		handleOnNegotiationNeeded func(signalID, peerConnectionID, cookieID *string),
	) (*peer.Peer, error)
	GetAllConnectionID() map[string][]string
	GetStates() map[string]string

	AddConnections(signalID *string)
	GetConnections(signalID *string) peer.Connections
	GetConnection(signalID, peerConnectionID *string) (*peer.Peer, error)
	RemoveConnection(signalID, peerConnectionID, cookieID *string) error
	RemoveConnections(signalID *string)

	UnRegister(peerConnectionID *string)
	UnRegisterVideo(peerConnectionID, trackID *string)
	UnRegisterAudio(peerConnectionID, trackID *string)

	AddVideoFwd(trackID, codec string)
	AddAudioFwd(trackID, codec string)

	// add the method add register, the given trackID could be null if register all
	Register(
		signalID,
		peerConnectionID *string,
		videoTrackIDs,
		audioTrackIDs []string,
		errHandler func(signalID, peerConnectionID, trackID *string, reason string),
	) error
	RegisterAudio(
		signalID,
		audioTrackID,
		peerConnectionID *string,
		errHandler func(signalID, peerConnectionID, trackID *string, reason string),
	) error
	RegisterVideo(
		signalID,
		videoTrackID,
		peerConnectionID *string,
		errHandler func(signalID, peerConnectionID, trackID *string, reason string),
	) error

	SetUpList(lst map[string]*UpPeer)
	DeleteUpList(peerConnectionID *string)
	AddUpList(peerConnectionID *string, c *UpPeer)
	AppendUpList(pcID *string, obj *UpPeer)

	GetRemoteTrack(trackID string) *webrtc.TrackRemote
	SetHandleNoConnection(handler func(signalID *string))

	GetVideoReceiveTime() map[string]int64
	GetAudioReceiveTime() map[string]int64
	GetVideoReceiveTimeby(trackID string) int64
	GetAudioReceiveTimeby(trackID string) int64

	GetTrackMeta(trackID string) bool
	SetTrackMeta(trackID string, state bool)
	SetHandleReadDeadline(f func(pcID, trackID *string, codec, kind string))
}
