package peer

import (
	"fmt"
	"sync"

	"github.com/lamhai1401/gologs/logs"
	"github.com/spgnk/rtc/errs"
	"github.com/spgnk/rtc/utils"

	"github.com/pion/webrtc/v3"
)

// LocalTracks control sending track
type LocalTracks struct {
	// mode         *string // mode to set default codec or modify
	conn           *webrtc.PeerConnection
	trans          map[string]*webrtc.RTPTransceiver
	videoTracks    map[string]webrtc.TrackLocal
	audioTracks    map[string]webrtc.TrackLocal
	videoSenders   map[string]*webrtc.RTPSender
	audioSenders   map[string]*webrtc.RTPSender
	receiveData    map[string]bool // save to trackID - state
	firstInitTrack map[string]string
	mutex          sync.RWMutex
}

// NewTracks linter
func NewTracks(
	conn *webrtc.PeerConnection,
) *LocalTracks {
	l := &LocalTracks{
		conn:           conn,
		trans:          make(map[string]*webrtc.RTPTransceiver),
		videoTracks:    make(map[string]webrtc.TrackLocal),
		audioTracks:    make(map[string]webrtc.TrackLocal),
		videoSenders:   make(map[string]*webrtc.RTPSender),
		audioSenders:   make(map[string]*webrtc.RTPSender),
		receiveData:    make(map[string]bool),
		firstInitTrack: make(map[string]string),
	}

	return l
}

// setCodecPreferences this only support for video because audio was fixed opus
func (t *LocalTracks) setCodecPreferences(payLoadType *int, trackConfig *TrackConfig) error {
	if payLoadType != nil {
		trans := t.getTrans(trackConfig.trackID)
		if trans == nil {
			return errs.ErrP006
		}

		err := trans.SetCodecPreferences(trackConfig.GetRTPCodecCapability(payLoadType))
		if err != nil {
			return err
		}

		return nil
	}
	return fmt.Errorf("payLoadType is nil")
}

// InitLocalTrack nil input payLoadType it mean get default
func (t *LocalTracks) createLocalVideo(trackConfig *TrackConfig) error {
	// remove old track
	t.deleteVideoTracks(trackConfig.trackID)
	var videoTrack webrtc.TrackLocal

	var err error
	if trackConfig.kind != nil && *trackConfig.kind == utils.SampleTrackType {
		videoTrack, err = t._createTrackSample(trackConfig.trackID, &trackConfig.codec)
		if err != nil {
			return err
		}
	} else {
		videoTrack, err = t._createTrack(trackConfig.trackID, &trackConfig.codec)
		if err != nil {
			return err
		}
	}

	// set track
	t.setVideoTracks(trackConfig.trackID, videoTrack)

	// trans, err := t._initTransceiver(videoTrack, trackConfig.role)
	// if err != nil {
	// 	return err
	// }

	sender, err := t.conn.AddTrack(videoTrack)
	if err != nil {
		return err
	}

	// remove sender
	t.deleteVideoSender(trackConfig.trackID)

	// remove transceiver
	t.closeTrans(trackConfig.trackID)

	// save sender video
	t.addVideoSender(trackConfig.trackID, sender)

	// save transceiver
	// t.setTrans(trackConfig.trackID, trans)

	go t._processRTCP(sender)

	// logs.Debug("Local video was created with config ")
	// if os.Getenv("DEBUG") == "1" {
	// 	spew.Dump(trackConfig)
	// }

	// set first init
	if t.getFirstInitTrack(trackConfig.trackID) == "" {
		t.setFirstInitTrack(trackConfig.trackID, "1")
	}
	return nil
}

// InitLocalTrack create opus audio
func (t *LocalTracks) createLocalAudio(trackConfig *TrackConfig) error {
	// remove old track
	t.deleteAudioTracks(trackConfig.trackID)

	opus := utils.ModeOpus
	var audioTrack webrtc.TrackLocal
	var err error

	if trackConfig.kind != nil && *trackConfig.kind == utils.SampleTrackType {
		audioTrack, err = t._createTrackSample(trackConfig.trackID, &opus)
		if err != nil {
			return err
		}
	} else {
		audioTrack, err = t._createTrack(trackConfig.trackID, &opus)
		if err != nil {
			return err
		}
	}
	// set track
	t.setAudioTracks(trackConfig.trackID, audioTrack)

	// trans, err := t._initTransceiver(audioTrack, trackConfig.role)
	// if err != nil {
	// 	return err
	// }

	sender, err := t.conn.AddTrack(audioTrack)
	if err != nil {
		return err
	}

	// remove sender
	t.deleteAudioSender(trackConfig.trackID)

	// remove transceiver
	// t.closeTrans(trackConfig.trackID)

	// save sender video
	t.addAudioSender(trackConfig.trackID, sender)

	// save transceiver
	// t.setTrans(trackConfig.trackID, trans)

	// logs.Debug("Local audio was created with config ")
	// if os.Getenv("DEBUG") == "1" {
	// 	spew.Dump(trackConfig)
	// }

	go t._processRTCP(sender)

	// set first init
	if t.getFirstInitTrack(trackConfig.trackID) == "" {
		t.setFirstInitTrack(trackConfig.trackID, "1")
	}
	return nil
}

func (t *LocalTracks) _processRTCP(rtpSender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, 1500)
	for {
		if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
			return
		}
	}
}

func (t *LocalTracks) removeLocalVideoTrack(trackID *string) error {
	// delete sender
	t.deleteVideoSender(trackID)

	// delete track
	t.deleteVideoTracks(trackID)

	// remove video receive track
	t.deleteReceiveData(trackID)

	// // get trans
	// trans := t.getTrans(trackID)
	// if trans == nil {
	// 	return fmt.Errorf("%s transceiver is nil", *trackID)
	// }

	// // delete trans
	// defer t.deleteTrans(trackID)

	// // close trans
	// err := trans.Stop()
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (t *LocalTracks) removeLocalAudioTrack(trackID *string) error {
	// delete sender
	t.deleteAudioSender(trackID)

	// delete track
	t.deleteAudioTracks(trackID)

	// remove audio receive track
	t.deleteReceiveData(trackID)

	// // get trans
	// trans := t.getTrans(trackID)
	// if trans == nil {
	// 	return fmt.Errorf("%s transceiver is nil", *trackID)
	// }

	// // delete trans
	// defer t.deleteTrans(trackID)

	// // close trans
	// err := trans.Stop()
	// if err != nil {
	// 	return err
	// }
	return nil
}

// CreateTrack linter
func (t *LocalTracks) _createTrack(id, codec *string) (*webrtc.TrackLocalStaticRTP, error) {
	return webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: utils.GetCodec(codec)},
		*id,
		*id,
	)
}

func (t *LocalTracks) _createTrackSample(id, codec *string) (*webrtc.TrackLocalStaticSample, error) {
	return webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: utils.GetCodec(codec)},
		*id,
		*id,
	)
}

func (t *LocalTracks) getAudioSender(id *string) *webrtc.RTPSender {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.audioSenders[*id]
}

func (t *LocalTracks) getVideoSender(id *string) *webrtc.RTPSender {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.videoSenders[*id]
}

func (t *LocalTracks) deleteAudioSender(id *string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.audioSenders, *id)
}

func (t *LocalTracks) deleteVideoSender(id *string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.videoSenders, *id)
}

func (t *LocalTracks) addAudioSender(id *string, sender *webrtc.RTPSender) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.audioSenders[*id] = sender
}

func (t *LocalTracks) addVideoSender(id *string, sender *webrtc.RTPSender) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.videoSenders[*id] = sender
}

func (t *LocalTracks) setVideoTracks(id *string, track webrtc.TrackLocal) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.videoTracks[*id] = track
}

func (t *LocalTracks) setAudioTracks(id *string, track webrtc.TrackLocal) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.audioTracks[*id] = track
}

func (t *LocalTracks) deleteAudioTracks(id *string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.audioTracks, *id)
}

func (t *LocalTracks) deleteVideoTracks(id *string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.videoTracks, *id)
}

func (t *LocalTracks) getVideoTrack(trackID *string) *webrtc.TrackLocalStaticRTP {
	var result *webrtc.TrackLocalStaticRTP
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if track := t.videoTracks[*trackID]; track != nil {
		t, ok := track.(*webrtc.TrackLocalStaticRTP)
		if ok {
			result = t
		}
	}
	return result
}

func (t *LocalTracks) getAudioTrack(trackID *string) *webrtc.TrackLocalStaticRTP {
	var result *webrtc.TrackLocalStaticRTP
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if track := t.audioTracks[*trackID]; track != nil {
		t, ok := track.(*webrtc.TrackLocalStaticRTP)
		if ok {
			result = t
		}
	}
	return result
}

// initTransceiver init peer transceiver
func (t *LocalTracks) _initTransceiver(track webrtc.TrackLocal, role *string) (*webrtc.RTPTransceiver, error) {
	var dir webrtc.RTPTransceiverDirection

	switch *role {
	case "up":
		dir = webrtc.RTPTransceiverDirectionSendonly
	case "down":
		dir = webrtc.RTPTransceiverDirectionSendrecv
	default:
		dir = webrtc.RTPTransceiverDirectionSendrecv
	}

	trans, err := t.conn.AddTransceiverFromTrack(track, webrtc.RTPTransceiverInit{Direction: dir})
	if err != nil {
		return nil, err
	}

	// trans.SetCodecPreferences(codecs []webrtc.RTPCodecParameters
	return trans, nil
}

func (t *LocalTracks) closeTrans(trackID *string) error {
	tran := t.getTrans(trackID)
	if tran == nil {
		return fmt.Errorf(*trackID, errs.ErrP006.Error())
	}

	err := tran.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (t *LocalTracks) deleteTrans(trackID *string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.trans, *trackID)
}

func (t *LocalTracks) setTrans(trackID *string, tran *webrtc.RTPTransceiver) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.trans[*trackID] = tran
}

func (t *LocalTracks) getTrans(trackID *string) *webrtc.RTPTransceiver {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.trans[*trackID]
}

func (t *LocalTracks) getAudioTrackSample(trackID *string) *webrtc.TrackLocalStaticSample {
	var result *webrtc.TrackLocalStaticSample
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if track := t.audioTracks[*trackID]; track != nil {
		t, ok := track.(*webrtc.TrackLocalStaticSample)
		if ok {
			result = t
		}
	}
	return result
}

func (t *LocalTracks) getVideoTrackSample(trackID *string) *webrtc.TrackLocalStaticSample {
	var result *webrtc.TrackLocalStaticSample
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if track := t.videoTracks[*trackID]; track != nil {
		t, ok := track.(*webrtc.TrackLocalStaticSample)
		if ok {
			result = t
		}
	}
	return result
}

func (t *LocalTracks) replaceVideoRTPTrack(trackID, codec *string) error {
	// remove old track
	rtpTrack := t.getVideoTrack(trackID)
	if rtpTrack == nil {
		return fmt.Errorf(*trackID, "old video track is nil")
	}
	t.deleteVideoTracks(trackID)

	// create new local track
	newRTPTrack, err := t._createTrack(trackID, codec)
	if err != nil {
		return err
	}

	// get sender
	sender := t.getVideoSender(trackID)
	if sender == nil {
		return fmt.Errorf("%s video sender is nil", *trackID)
	}

	// new track
	err = sender.ReplaceTrack(newRTPTrack)
	if err != nil {
		return err
	}

	t.setVideoTracks(trackID, newRTPTrack)
	logs.Info(*trackID, "has new video track with mimtype", *codec)
	return nil
}

func (t *LocalTracks) replaceAudioRTPTrack(trackID, codec *string) error {
	// remove old track
	rtpTrack := t.getAudioTrack(trackID)
	if rtpTrack == nil {
		return fmt.Errorf(*trackID, "old audio track is nil")
	}
	t.deleteAudioTracks(trackID)

	// create new local track
	newRTPTrack, err := t._createTrack(trackID, codec)
	if err != nil {
		return err
	}

	// get sender
	sender := t.getAudioSender(trackID)
	if sender == nil {
		return fmt.Errorf("%s audio sender is nil", *trackID)
	}

	// new track
	err = sender.ReplaceTrack(newRTPTrack)
	if err != nil {
		return err
	}

	t.setAudioTracks(trackID, newRTPTrack)
	logs.Info(*trackID, "has new audio track with mimtype", *codec)
	return nil
}

func (t *LocalTracks) getReceiveData(trackID *string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.receiveData[*trackID]
}

func (t *LocalTracks) setReceiveData(trackID *string, state bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.receiveData[*trackID] = state
}

func (t *LocalTracks) deleteReceiveData(trackID *string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.receiveData, *trackID)
}

func (t *LocalTracks) setFirstInitTrack(trackID *string, state string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.firstInitTrack[*trackID] = state
}

func (t *LocalTracks) getFirstInitTrack(trackID *string) string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.firstInitTrack[*trackID]
}
