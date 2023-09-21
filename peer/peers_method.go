package peer

import (
	"fmt"

	"github.com/spgnk/rtc/utils"
)

func (p *Peers) getSignalID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return *p.signalID
}

func (p *Peers) setClosed(state bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isClosed = state
}

func (p *Peers) wasClosed() bool {
	return p.isClosed
}

func (p *Peers) getAudioFwdm() *utils.ForwarderMannager {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.audioFwdm
}

func (p *Peers) getVideoFwdm() *utils.ForwarderMannager {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.videoFwdm
}

func (p *Peers) getPeers() *utils.AdvanceMap {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.peers
}

func (p *Peers) setPeer(peerConnectionID *string, c *Peer) {
	if p := p.getPeers(); p != nil {
		p.Set(*peerConnectionID, c)
	}
}

func (p *Peers) deletePeer(peerConnectionID *string) {
	if clients := p.getPeers(); clients != nil {
		clients.Delete(*peerConnectionID)
	}
}

func (p *Peers) getPeer(peerConnectionID *string) *Peer {
	if clients := p.getPeers(); clients != nil {
		c, ok := clients.Get(*peerConnectionID)
		if !ok {
			return nil
		}
		client, ok := c.(*Peer)
		if ok {
			return client
		}
	}
	return nil
}

func (p *Peers) closePeer(peerConnectionID *string) {
	if client := p.getPeer(peerConnectionID); client != nil {
		p.deletePeer(peerConnectionID)

		// remove from fwdm
		if fwdm := p.getVideoFwdm(); fwdm != nil {
			fwdm.UnregisterAll(*client.GetPeerConnectionID())
		}
		if fwdm := p.getAudioFwdm(); fwdm != nil {
			fwdm.UnregisterAll(*client.GetPeerConnectionID())
		}

		// close peer
		client.Close()

		p.logger.INFO(fmt.Sprintf("%s_%s peerConn was removed", *peerConnectionID, *client.getCookieID()), nil)
		client = nil
	}
}

func (p *Peers) getStates() *utils.AdvanceMap {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.states
}

func (p *Peers) getState(peerConnectionID string) string {
	if states := p.getStates(); states != nil {
		state, has := states.Get(peerConnectionID)
		if !has {
			return ""
		}
		s, ok := state.(string)
		if ok {
			return s
		}
	}

	return ""
}

func (p *Peers) setState(peerConnectionID, state *string) {
	if states := p.getStates(); states != nil {
		states.Set(*peerConnectionID, *state)
	}
}

func (p *Peers) deleteState(peerConnectionID *string) {
	if states := p.getStates(); states != nil {
		states.Delete(*peerConnectionID)
	}
}

// exportStates iter state and export into a map string
func (p *Peers) exportStates(temp map[string]string) {
	if states := p.getStates(); states != nil {
		states.Iter(func(key, value interface{}) bool {
			k, ok1 := key.(string)
			v, ok2 := value.(string)
			if ok1 && ok2 {
				temp[k] = v
			}
			return true
		})
	}
}

// RemoveConnections remove all connection
func (p *Peers) removeConnections() {
	if peers := p.getPeers(); peers != nil {
		keys := peers.GetKeys()
		for _, key := range keys {
			p.RemoveConnection(&key)
		}
	}
}

func (p *Peers) getAllPeer() ([]*Peer, []string) {
	var conns []*Peer
	var ids []string

	if ps := p.getPeers(); ps != nil {
		ps.Iter(func(key, value interface{}) bool {
			connection, ok1 := value.(*Peer)
			if ok1 {
				conns = append(conns, connection)
			}

			streamID, ok2 := key.(string)
			if ok2 {
				ids = append(ids, streamID)
			}
			return true
		})
	}

	return conns, ids
}
