package peer

import (
	"fmt"
	"strings"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
	"github.com/spgnk/rtc/utils"
)

// InitAPI linter
func (p *Peer) initAPI(config *Configs) (*webrtc.API, error) {
	// init media engine
	m, err := p.initMediaEngine(config)
	if err != nil {
		return nil, err
	}
	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		// logs.Error("initAPI RegisterDefaultInterceptors error: ", err.Error())
		return nil, err
	}

	return webrtc.NewAPI(
		webrtc.WithMediaEngine(m),
		webrtc.WithInterceptorRegistry(i),
		webrtc.WithSettingEngine(*p.initSettingEngine(config))), nil
}

func (p *Peer) initSettingEngine(config *Configs) *webrtc.SettingEngine {
	settingEngine := &webrtc.SettingEngine{}

	var mtu int

	if config.MaxMTU == 0 {
		mtu = config.MaxMTU
	}

	// SetReceiveMTU sets the size of read buffer that copies incoming packets. This is optional.
	settingEngine.SetReceiveMTU(uint(mtu))

	// settingEngine.SetEphemeralUDPPortRange(20000, 60000)
	// settingEngine.SetICETimeouts(10*time.Second, 20*time.Second, 1*time.Second)
	return settingEngine
}

func (p *Peer) initMediaEngine(config *Configs) (*webrtc.MediaEngine, error) {
	mediaEngine := &webrtc.MediaEngine{}

	videoRTCPFeedback := []webrtc.RTCPFeedback{
		{Type: "goog-remb", Parameter: ""},
		{Type: "ccm", Parameter: "fir"},
		{Type: "nack", Parameter: ""},
		{Type: "nack", Parameter: "pli"},
	}

	switch *config.Role {
	case utils.PeerDown:
		if config.Codec != nil {
			// register video profile
			switch strings.ToLower(*config.Codec) {
			case utils.ModeVP8:
				if config.PayloadType == 0 {
					config.PayloadType = utils.DefaultPayloadVP8
				}
				err := p.registerVP8(mediaEngine, config.PayloadType, videoRTCPFeedback)
				if err != nil {
					return nil, err
				}
				// register opus
				err = p.registerOpus(mediaEngine)
				if err != nil {
					return nil, err
				}
			case utils.ModeVP9:
				if config.PayloadType == 0 {
					config.PayloadType = utils.DefaultPayloadVP9
				}
				err := p.registerVP9(mediaEngine, config.PayloadType, videoRTCPFeedback)
				if err != nil {
					return nil, err
				}
				// register opus
				err = p.registerOpus(mediaEngine)
				if err != nil {
					return nil, err
				}
			case utils.ModeH264:
				err := p.registerH264(mediaEngine, config.PayloadType, videoRTCPFeedback)
				if err != nil {
					return nil, err
				}
				// register opus
				err = p.registerOpus(mediaEngine)
				if err != nil {
					return nil, err
				}
			default:
				err := mediaEngine.RegisterDefaultCodecs()
				if err != nil {
					return nil, err
				}
			}
		} else {
			err := mediaEngine.RegisterDefaultCodecs()
			if err != nil {
				return nil, err
			}
		}
	case utils.PeerUp:
		err := mediaEngine.RegisterDefaultCodecs()
		if err != nil {
			return nil, err
		}
	default:
		err := mediaEngine.RegisterDefaultCodecs()
		if err != nil {
			return nil, err
		}
	}
	return mediaEngine, nil
}

func (p *Peer) registerVP8(m *webrtc.MediaEngine, payload int, videoRTCPFeedback []webrtc.RTCPFeedback) error {
	err := p._addVP8(m, payload, videoRTCPFeedback)
	if err != nil {
		return err
	}

	// add extensions
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     "video/ulpfec",
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
		PayloadType: 116,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}
	// Default Pion Video Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) registerVP9(m *webrtc.MediaEngine, payload int, videoRTCPFeedback []webrtc.RTCPFeedback) error {
	profileID := 0
	err := p._addVP9(m, payload, videoRTCPFeedback, &profileID)
	if err != nil {
		return err
	}
	err = p._addVP9(m, payload+2, videoRTCPFeedback, &profileID)
	if err != nil {
		return err
	}

	// add extensions
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     "video/ulpfec",
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
		PayloadType: 116,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}
	// Default Pion Video Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) registerOpus(m *webrtc.MediaEngine) error {
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: utils.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1", RTCPFeedback: nil},
			PayloadType:        111,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: utils.MimeTypeG722, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        9,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: utils.MimeTypePCMU, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        0,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: utils.MimeTypePCMA, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        8,
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
	}

	// Default Pion Audio Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) _addVP8(m *webrtc.MediaEngine, payload int, videoRTCPFeedback []webrtc.RTCPFeedback) error {
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeVP8,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "",
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: webrtc.PayloadType(payload),
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/rtx",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("apt=%d", payload),
				RTCPFeedback: nil,
			},
			PayloadType: webrtc.PayloadType(payload + 1),
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) _addVP9(m *webrtc.MediaEngine, payload int, videoRTCPFeedback []webrtc.RTCPFeedback, profileID *int) error {
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeVP9,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("profile-id=%d", *profileID),
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: webrtc.PayloadType(payload),
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/rtx",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("apt=%d", payload),
				RTCPFeedback: nil,
			},
			PayloadType: webrtc.PayloadType(payload + 1),
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}
	*profileID++
	return nil
}

func (p *Peer) registerH264(m *webrtc.MediaEngine, payload int, videoRTCPFeedback []webrtc.RTCPFeedback) error {
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeH264,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: 102,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/rtx",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "apt=102",
				RTCPFeedback: nil},
			PayloadType: 121,
		},

		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeH264,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f",
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: 127,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/rtx",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "apt=127",
				RTCPFeedback: nil,
			},
			PayloadType: 120,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeH264,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f",
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: 127,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/rtx",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "apt=127",
				RTCPFeedback: nil,
			},
			PayloadType: 120,
		},

		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     utils.MimeTypeH264,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=640032",
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: 123,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/rtx",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "apt=123",
				RTCPFeedback: nil,
			},
			PayloadType: 118,
		},

		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/ulpfec",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "",
				RTCPFeedback: nil,
			},
			PayloadType: 116,
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}

	// Default Pion Video Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}

	return nil
}
