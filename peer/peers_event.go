package peer

import (
	"fmt"
	"time"

	"github.com/spgnk/rtc/errs"
	"github.com/spgnk/rtc/utils"
)

func (p *Peers) clearDisconnected(pcID *string) {
	id := fmt.Sprintf("disconnected-%s", *pcID)
	utils.RemoveTask(id)
}

func (p *Peers) checkDisconnected(pcID *string, f func()) {
	id := fmt.Sprintf("disconnected-%s", *pcID)
	utils.AddTask(id, time.AfterFunc(10000*time.Millisecond, f))
}

func (p *Peers) handleICEConnectionState(
	signalID, peerConnectionID *string, state string,
	handleAddPeer func(signalID, role, peerConnectionID *string),
	handleFailedPeer func(signalID, role, peerConnectionID *string),
) {
	peer := p.getPeer(peerConnectionID)
	if peer == nil || state == "" {
		return
	}
	peer.Info(
		fmt.Sprintf("%s_%s current ICE states: %s", *signalID, *peer.getCookieID(), state),
		map[string]any{
			"signal_id": p.getSignalID(),
		})
	p.setState(peerConnectionID, &state)

	switch state {
	case utils.Connected:
		p.clearDisconnected(peerConnectionID)
		if !peer.IsConnected() { // notif if this is new peer
			// remove ice cache
			peer.clearIceCache()
			// setting connected, call handler
			peer.SetIsConnected(true)
			if handleAddPeer != nil {
				handleAddPeer(signalID, peer.getRole(), peer.GetPeerConnectionID())
			}
		}
	case utils.Disconnected:
		p.checkDisconnected(peerConnectionID, func() {
			p.checkFailedState(peer.GetPeerConnectionID(), peer.getCookieID(), handleFailedPeer)
		})
	case utils.Failed:
		// utils.AddTask(*peer.getCookieID(), time.AfterFunc(5*time.Second, func() {
		// 	p.checkFailedState(peer.GetPeerConnectionID(), peer.getCookieID(), handleFailedPeer)
		// }))
		p.checkFailedState(peer.GetPeerConnectionID(), peer.getCookieID(), handleFailedPeer)
	case utils.Closed:
		peer.Info(fmt.Sprintf("rtcn ICE state is %s", state), map[string]any{
			"signal_id": p.getSignalID(),
		})
	default:
		return
	}
}

func (p *Peers) checkFailedState(
	peerConnectionID, cookieID *string,
	handleFailedPeer func(signalID, role, peerConnectionID *string),
) {
	state := p.getState(*peerConnectionID)
	peer := p.GetConnection(peerConnectionID)
	if peer == nil {
		return
	}
	if (state == "failed" || state == "disconnected" || state == "closed") && *peerConnectionID == *peer.GetPeerConnectionID() && *cookieID == *peer.GetCookieID() {
		p.Error(errs.ErrP004.Error())
		p.RemoveConnection(peerConnectionID)
		p.Warn(fmt.Sprintf("Remove old peerConn (%s_%s_%s) has state %s", p.getSignalID(), *peerConnectionID, *peer.GetCookieID(), state))
		if handleFailedPeer != nil {
			handleFailedPeer(p.signalID, peer.GetRole(), peer.GetPeerConnectionID())
		}
	}
}
