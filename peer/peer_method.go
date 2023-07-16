package peer

import (
	"fmt"
	"time"

	"github.com/spgnk/rtc/errs"
	"github.com/spgnk/rtc/utils"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

func (p *Peer) stackDebug(v ...string) {
	if p.debug == "1" {
		p.Stack(v...)
	}
}

func (p *Peer) checkClose() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isClosed
}

func (p *Peer) setClose(state bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isClosed = state
}

func (p *Peer) writeRTP(packet *rtp.Packet, track *webrtc.TrackLocalStaticRTP) error {
	return track.WriteRTP(packet)
}

func (p *Peer) writeSample(sample *media.Sample, track *webrtc.TrackLocalStaticSample) error {
	return track.WriteSample(*sample)
}

// SetBitrate linter
func (p *Peer) setBitrate(bitrate *int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.bitrate = bitrate
}

func (p *Peer) getCookieID() *string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.cookieID
}

func (p *Peer) getBitrate() *int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.bitrate
}

func (p *Peer) getRole() *string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.config.Role
}

func (p *Peer) setConn(conn *webrtc.PeerConnection) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.conn = conn
}

func (p *Peer) getConn() *webrtc.PeerConnection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.conn
}

func (p *Peer) closeConn() error {
	if conn := p.getConn(); conn != nil {
		// remove task if has
		utils.RemoveTask(*p.getCookieID())

		p.setConn(nil)
		err := conn.Close()
		if err != nil {
			return err
		}
		conn = nil
	}
	return nil
}

// GetLocalDescription get current peer local description
func (p *Peer) getLocalDescription() (*webrtc.SessionDescription, error) {
	conn := p.getConn()
	if conn == nil {
		return nil, errs.ErrP002
	}
	return conn.LocalDescription(), nil
}

func (p *Peer) getIceCache() *utils.AdvanceMap {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.iceCache
}

func (p *Peer) addIceCache(ice *webrtc.ICECandidateInit) {
	if cache := p.getIceCache(); cache != nil {
		cache.Set(ice.Candidate, ice)
	}
}

// setCacheIce add ice save in cache
func (p *Peer) setCacheIce() error {
	cache := p.getIceCache()
	if cache == nil {
		return fmt.Errorf("ICE cache map is nil")
	}
	conn := p.getConn()
	if conn == nil {
		return errs.ErrP002
	}

	captureCache := cache.Capture()
	for _, value := range captureCache {
		// ice, ok := value.(*webrtc.ICECandidateInit)
		// if ok {
		// 	if err := p.AddICECandidate(ice); err != nil {
		// 		return err
		// 	}
		// }
		if err := p.AddICECandidate(value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) clearIceCache() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.iceCache = utils.NewAdvanceMap()
}

// ModifyBitrate so set bitrate when datachannel has signal
// Use this only for video not audio track
func (p *Peer) modifyBitrate(remoteTrack *webrtc.TrackRemote) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for range ticker.C {
		bitrate := p.getBitrate()
		if p.checkClose() || bitrate == nil {
			return
		}

		numbers := *bitrate * 1024
		if conn := p.getConn(); conn != nil {
			errSend := conn.WriteRTCP([]rtcp.Packet{&rtcp.ReceiverEstimatedMaximumBitrate{
				SenderSSRC: uint32(remoteTrack.SSRC()),
				Bitrate:    float32(numbers),
				// SSRCs:      []uint32{rand.Uint32()},
			}})

			if errSend != nil {
				p.Error("Modify bitrate write rtcp err: " + errSend.Error())
				return
			}
		}
	}
}

// PictureLossIndication packet informs the encoder about the loss of an undefined amount of coded video data belonging to one or more pictures
func (p *Peer) pictureLossIndication(remoteTrack *webrtc.TrackRemote) {
	interval := 30
	if p.pli != 0 {
		interval = p.pli
	}

	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for range ticker.C {
		if p.checkClose() {
			return
		}

		conn := p.getConn()
		if conn == nil {
			return
		}
		errSend := conn.WriteRTCP([]rtcp.Packet{
			&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
			// &rtcp.SliceLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
			// &rtcp.RapidResynchronizationRequest{SenderSSRC: uint32(remoteTrack.SSRC()), MediaSSRC: uint32(remoteTrack.SSRC())},
		})

		if errSend != nil {
			p.Error("Picture loss indication write rtcp err: " + errSend.Error())
			return
		}
	}
}

// SendPictureLossIndication linter
func (p *Peer) SendPictureLossIndication() {
	remoteTrack := p.getRemoteTrack()
	if remoteTrack == nil {
		return
	}
	conn := p.getConn()
	if conn == nil {
		return
	}
	errSend := conn.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
		// &rtcp.SliceLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
		// &rtcp.RapidResynchronizationRequest{SenderSSRC: uint32(remoteTrack.SSRC()), MediaSSRC: uint32(remoteTrack.SSRC())},
	})

	if errSend != nil {
		p.Error("Picture loss indication write rtcp err: " + errSend.Error())
		return
	}
}

func (p *Peer) getRemoteTrack() *webrtc.TrackRemote {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.remoteTrack
}

func (p *Peer) setRemoteTrack(t *webrtc.TrackRemote) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.remoteTrack = t
}

func (p *Peer) setDuplicated(t string, state bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.duplicated[t] = state
}

// DeleteDuplicated linter
func (p *Peer) DeleteDuplicated(t string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.duplicated, t)
}

// GetDuplicated check duplicated
func (p *Peer) GetDuplicated(t string) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.duplicated[t]
}

// AddDuplicated to add and automatic delete after 10s
func (p *Peer) AddDuplicated(t string, element bool) {
	p.setDuplicated(t, element)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		<-ticker.C
		if p.GetDuplicated(t) {
			p.DeleteDuplicated(t)
		}
	}()
}
