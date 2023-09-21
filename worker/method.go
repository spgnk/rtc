package worker

import (
	"fmt"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/spgnk/rtc/peer"
	"github.com/spgnk/rtc/utils"
)

func (w *PeerWorker) getPeers() *utils.AdvanceMap {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.peers
}

func (w *PeerWorker) countAllPeer() {
	var all int64
	if peers := w.getPeers(); peers != nil {
		peers.Iter(func(key, value interface{}) bool {
			signalID, ok1 := key.(string)
			connections, ok2 := value.(peer.Connections)
			if ok1 && ok2 {
				count := connections.CountAllPeer()
				if count == 0 {
					// something here
					w.logger.WARN(fmt.Sprintf("==== %s has 0 connection. Check repeer or remove", signalID), nil)
					if w.handleNoConnection != nil {
						w.handleNoConnection(&signalID)
					}
				} else {
					all += count
					w.logger.WARN(fmt.Sprintf("==== %s has %d connections", signalID, count), nil)
				}
			}

			return true
		})
	}

	w.logger.WARN(fmt.Sprintf("==== Total connections is: %d", all), nil)
}

func (w *PeerWorker) countInterVal() {
	interval := utils.GetInterval()
	w.logger.WARN(fmt.Sprintf("Count interval start with %d every second", interval), nil)
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for range ticker.C {
		w.logger.WARN("====== Count peer interval ======", nil)
		w.countAllPeer()
	}
}

func (w *PeerWorker) deleteConnections(signalID *string) {
	if ps := w.getPeers(); ps != nil {
		ps.Delete(*signalID)
	}
}

func (w *PeerWorker) getConnections(signalID *string) peer.Connections {
	if ps := w.getPeers(); ps != nil {
		connections, has := ps.Get(*signalID)
		if has {
			conns, ok := connections.(peer.Connections)
			if ok {
				return conns
			}
		}
	}
	return nil
}

func (w *PeerWorker) closeConnections(signalID *string) {
	if conns := w.getConnections(signalID); conns != nil {
		w.deleteConnections(signalID)
		conns.Close()
		conns = nil
	}
}

func (w *PeerWorker) getPeer(signalID, peerConnectionID *string) *peer.Peer {
	if peers := w.getPeers(); peers != nil {
		ps, has := peers.Get(*signalID)
		if has {
			connections, ok := ps.(peer.Connections)
			if ok {
				return connections.GetConnection(peerConnectionID)
			}
		}
	}
	return nil
}

func (w *PeerWorker) deleteUpList(peerConnectionID *string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.upList, *peerConnectionID)
}

func (w *PeerWorker) addUpList(peerConnectionID *string, c *UpPeer) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.upList[*peerConnectionID] = c
}

func (w *PeerWorker) getUpList() map[string]*UpPeer {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.upList
}

func (w *PeerWorker) setUpList(lst map[string]*UpPeer) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.upList = lst
}

func (w *PeerWorker) getUpPeer(pcID *string) *UpPeer {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.upList[*pcID]
}

func (w *PeerWorker) setUpPeer(pcID *string, obj *UpPeer) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.upList[*pcID] = obj
}

// func (w *PeerWorker) getAudioFwdm() utils.Fwdm {
// 	w.mutex.RLock()
// 	defer w.mutex.RUnlock()
// 	return w.audioFwdm
// }

// func (w *PeerWorker) getVideoFwdm() utils.Fwdm {
// 	w.mutex.RLock()
// 	defer w.mutex.RUnlock()
// 	return w.videoFwdm
// }

func (w *PeerWorker) setRemoteTrack(trackID string, track *webrtc.TrackRemote) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.tracks[trackID] = track
}

func (w *PeerWorker) getRemoteTrack(trackID string) *webrtc.TrackRemote {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.tracks[trackID]
}

func (w *PeerWorker) deleteRemoteTrack(trackID string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.tracks, trackID)
}
