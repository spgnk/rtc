package peer

import (
	"fmt"
	"sync"

	"github.com/lamhai1401/gologs/loki"
	"github.com/spgnk/rtc/errs"
	"github.com/spgnk/rtc/utils"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// Peers handle mutilpe peer connection
type Peers struct {
	signalID  *string
	audioFwdm utils.Fwdm             // forward audio pkg
	videoFwdm utils.Fwdm             // forward video pkg
	states    *utils.AdvanceMap      // save all client state peerConnectionID - state
	peers     *utils.AdvanceMap      // save peerConnectionID - peer
	headers   map[string]*rtp.Header // save data header with for data header
	isClosed  bool
	mutex     sync.RWMutex
	logger    loki.Log
}

// NewPeers mutilpe peer controller
func NewPeers(
	signalID *string, // client signal ID
	logger loki.Log,
) Connections {
	ps := &Peers{
		signalID: signalID,
		peers:    utils.NewAdvanceMap(),
		isClosed: false,
		states:   utils.NewAdvanceMap(),
		headers:  make(map[string]*rtp.Header),
		logger:   logger,
	}

	// ps.serve()
	return ps
}

// Close linter
func (p *Peers) Close() {
	if !p.wasClosed() {
		p.setClosed(true)
		p.removeConnections()
	}
}

// GetConnection get peer connection
func (p *Peers) GetConnection(peerConnectionID *string) Connection {
	peer := p.getPeer(peerConnectionID)
	if peer != nil {
		return peer
	}
	return nil
}

// RemoveConnection remove existing connection
func (p *Peers) RemoveConnection(peerConnectionID *string) {
	p.closePeer(peerConnectionID)
	p.deleteState(peerConnectionID)
}

// GetState return a single peer connection state
func (p *Peers) GetState(peerConnectionID string) string {
	return p.getState(peerConnectionID)
}

// GetStates return all connection state
func (p *Peers) GetStates(temp map[string]string) {
	p.exportStates(temp)
}

// CountAllPeer count all existing peer
func (p *Peers) CountAllPeer() int64 {
	if peers := p.getPeers(); peers != nil {
		return peers.Len()
	}
	return 0
}

// AddDCConnection linter
func (p *Peers) AddDCConnection(
	configs *Configs,
	handleOnDatachannel func(pcID *string, d *webrtc.DataChannel),
	handleAddDCPeer func(signalID, role, peerConnectionID *string),
	handleFailedDCPeer func(signalID, role, peerConnectionID *string),
	handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
) (Connection, error) {
	// remove if exist
	if peer := p.GetConnection(configs.PeerConnectionID); peer != nil {
		p.RemoveConnection(configs.PeerConnectionID)
		p.Info(fmt.Sprintf("%s remove existed peerConn", *configs.PeerConnectionID))
	}

	// add new one
	configs.logger = p.logger
	peer := newPeerConnection(configs)

	conn, err := peer.InitDC(handleOnDatachannel)
	if err != nil {
		return nil, err
	}

	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil && handleCandidate != nil {
			handleCandidate(p.getSignalID(), configs.PeerConnectionID, candidate)
		}
	})

	conn.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		p.handleICEConnectionState(
			p.getSignalID(),
			configs.PeerConnectionID,
			is.String(),
			handleAddDCPeer,
			handleFailedDCPeer,
		)
	})

	p.setPeer(configs.PeerConnectionID, peer)
	return peer, nil
}

// AddConnection add new peer connection
func (p *Peers) AddConnection(
	configs *Configs,
	handleOnTrack func(signalID, peerConnectionID *string, remoteTrack *webrtc.TrackRemote),
	handleAddPeer func(signalID, role, peerConnectionID *string),
	handleFailedPeer func(signalID, role, peerConnectionID *string),
	handleCandidate func(signalID, peerConnectionID *string, candidate *webrtc.ICECandidate),
	handleOnNegotiationNeeded func(signalID, peerConnectionID, cookieID *string),
) (Connection, error) {
	// remove if exist
	if peer := p.GetConnection(configs.PeerConnectionID); peer != nil {
		p.RemoveConnection(configs.PeerConnectionID)
		p.Info(fmt.Sprintf("%s remove existed peerConn", *configs.PeerConnectionID))
	}

	// add new one
	configs.logger = p.logger
	peer := newPeerConnection(configs)

	conn, err := peer.Init()
	if err != nil {
		return nil, err
	}

	conn.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		kind := t.Kind().String()
		// find trackId in stream ọbject
		memDestCanPush := (kind == "video" && peer.config.AllowUpVideo) || (kind == "audio" && peer.config.AllowUpAudio)
		if !memDestCanPush {
			p.Warn(fmt.Sprintf("%s is not allow to up %s. No need to read RTP", *peer.GetPeerConnectionID(), kind))
			return
		}
		if kind == "video" {
			peer.HandleVideoTrack(t)
		}

		if handleOnTrack != nil {
			handleOnTrack(p.getSignalID(), peer.GetPeerConnectionID(), t)
		}
	})

	conn.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		p.handleICEConnectionState(
			p.getSignalID(),
			configs.PeerConnectionID,
			is.String(),
			handleAddPeer,
			handleFailedPeer,
		)
	})

	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil && handleCandidate != nil {
			handleCandidate(p.getSignalID(), configs.PeerConnectionID, candidate)
		}
	})

	// conn.OnNegotiationNeeded(func() {
	// 	if handleOnNegotiationNeeded != nil {
	// 		handleOnNegotiationNeeded(p.getSignalID(), configs.PeerConnectionID, peer.cookieID)
	// 	}
	// })

	p.setPeer(configs.PeerConnectionID, peer)
	return peer, nil
}

// AddCandidate linter
func (p *Peers) AddCandidate(peerConnectionID *string, value interface{}) error {
	conn := p.getPeer(peerConnectionID)
	if conn == nil {
		return errs.ErrPS003
	}

	// add candidate
	err := conn.AddICECandidate(value)
	if err != nil {
		return err
	}

	return nil
}

// AddSDP linter
func (p *Peers) AddSDP(peerConnectionID *string, value interface{}) error {
	conn := p.getPeer(peerConnectionID)
	if conn == nil {
		return errs.ErrP002
	}
	return conn.AddSDP(value)
}

// GetAllConnection return all connection
func (p *Peers) GetAllConnection() []Connection {
	tmp, _ := p.getAllPeer()
	return tmp
}

// GetAllPeerConnectionID return all streamID
func (p *Peers) GetAllPeerConnectionID() []string {
	_, tmp := p.getAllPeer()
	return tmp
}
