package peer

import (
	"fmt"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/spgnk/rtc/utils"
)

// Configs setting init peer config
type Configs struct {
	MaxMTU     int                   // setting for max buffer
	TurnConfig *webrtc.Configuration // turn server

	Codec *string
	// peer init
	Bitrate          *int
	Role             *string
	PeerConnectionID *string
	PayloadType      int
	AllowUpVideo     bool
	AllowUpAudio     bool
	AllowDownVideo   bool
	AllowDownAudio   bool

	IsCreateDC bool

	// Logger int // init logger
}

// TrackConfig setting init track config
type TrackConfig struct {
	kind              *string // rtp or sample track
	trackID           *string
	codec             string  // vp8/vp9/h264
	profileID         int     // for codec profile id
	role              *string // up or down
	videoRTCPFeedback []webrtc.RTCPFeedback
}

// NewTrackConfig linter
func NewTrackConfig(
	trackID *string,
	codec string, // vp8/vp9/h264
	role *string,
	kind *string,
) *TrackConfig {
	t := &TrackConfig{
		kind:      kind,
		role:      role,
		profileID: 0,
		trackID:   trackID,
		codec:     strings.ToLower(codec),
		videoRTCPFeedback: []webrtc.RTCPFeedback{
			{Type: "goog-remb", Parameter: ""},
			{Type: "ccm", Parameter: "fir"},
			{Type: "nack", Parameter: ""},
			{Type: "nack", Parameter: "pli"}},
	}

	return t
}

// GetRTPCodecCapability for add setCodecPreferences
func (t *TrackConfig) GetRTPCodecCapability(payloadType *int) []webrtc.RTPCodecParameters {
	switch t.codec {
	case utils.ModeVP8:
		return t.getVP8RTPCodecCapability(payloadType)
	case utils.ModeVP9:
		return t.getVP9RTPCodecCapability(payloadType)
	default:
		return t.getVP9RTPCodecCapability(payloadType)
	}
}

func (t *TrackConfig) getVP8RTPCodecCapability(payloadType *int) []webrtc.RTPCodecParameters {
	defer func() {
		// *payloadType += 2 // uncmt this line if rtx is on
	}()
	return []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeVP8,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "",
				RTCPFeedback: t.videoRTCPFeedback,
			},
			PayloadType: webrtc.PayloadType(*payloadType),
		},
		// {
		// 	RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/rtx", ClockRate: 90000, Channels: 0, SDPFmtpLine: fmt.Sprintf("apt=%d", payloadType), RTCPFeedback: nil},
		// 	PayloadType:        webrtc.PayloadType(payloadType + 1),
		// },
	}
}

func (t *TrackConfig) getVP9RTPCodecCapability(payloadType *int) []webrtc.RTPCodecParameters {
	defer func() {
		t.profileID++
		// *payloadType += 2 // uncmt this line if rtx is on
	}()
	return []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeVP9,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("profile-id=%d", t.profileID),
				RTCPFeedback: t.videoRTCPFeedback},
			PayloadType: webrtc.PayloadType(*payloadType),
		},
		// {
		// 	RTPCodecCapability: webrtc.RTPCodecCapability{
		// 		MimeType:     "video/rtx",
		// 		ClockRate:    90000,
		// 		Channels:     0,
		// 		SDPFmtpLine:  fmt.Sprintf("apt=%d", t.payloadType),
		// 		RTCPFeedback: nil,
		// 	},
		// 	PayloadType: webrtc.PayloadType(t.payloadType + 1),
		// },
	}
}
