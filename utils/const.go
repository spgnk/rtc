package utils

// const splitStr = "-"

const (
	// MimeTypeH264 H264 MIME type.
	// Note: Matching should be case insensitive.
	MimeTypeH264 = "video/H264"
	// MimeTypeOpus Opus MIME type
	// Note: Matching should be case insensitive.
	MimeTypeOpus = "audio/opus"
	// MimeTypeVP8 VP8 MIME type
	// Note: Matching should be case insensitive.
	MimeTypeVP8 = "video/VP8"
	// MimeTypeVP9 VP9 MIME type
	// Note: Matching should be case insensitive.
	MimeTypeVP9 = "video/VP9"
	// MimeTypeG722 G722 MIME type
	// Note: Matching should be case insensitive.
	MimeTypeG722 = "audio/G722"
	// MimeTypePCMU PCMU MIME type
	// Note: Matching should be case insensitive.
	MimeTypePCMU = "audio/PCMU"
	// MimeTypePCMA PCMA MIME type
	// Note: Matching should be case insensitive.
	MimeTypePCMA = "audio/PCMA"
)

// ice state
const (
	Connected    = "connected"
	Failed       = "failed"
	Closed       = "closed"
	Disconnected = "disconnected"
)

// MaxBufferSize The maximum amount of data that can be buffered before returning errors.
const MaxBufferSize = 1000 * 2048 // 1MB
// MaxMTU  linter
const MaxMTU = 1460 // 1MB

// Mediaenine mode
const (
	// ModeAll linter
	ModeAll = "all"
	// ModeVP8 liner
	ModeVP8 = "vp8"
	// ModeVP9 linter
	ModeVP9 = "vp9"
	// ModeH264 linter
	ModeH264 = "h264"
	// ModeOpus linter
	ModeOpus = "opus"
)

// GetModeType get mimetype codec
func GetModeType(codec string) string {
	switch codec {
	case MimeTypeVP8:
		return ModeVP8
	case MimeTypeVP9:
		return ModeVP9
	case MimeTypeH264:
		return ModeH264
	case MimeTypeOpus:
		return ModeOpus
	default:
		return ModeVP9
	}
}

// GetCodec get mimetype codec
func GetCodec(codec *string) string {
	switch *codec {
	case ModeVP8:
		return MimeTypeVP8
	case ModeVP9:
		return MimeTypeVP9
	case ModeH264:
		return MimeTypeH264
	case ModeOpus:
		return MimeTypeOpus
	default:
		return MimeTypeVP9
	}
}

var (
	// DefaultCodecVP8 linter
	DefaultCodecVP8 = "VP8"
	// DefaultCodecVP9 linter
	DefaultCodecVP9 = "VP9"
	// DefaultPayloadVP8 linter
	DefaultPayloadVP8 = 96
	// DefaultPayloadVP9 linter
	DefaultPayloadVP9 = 98
)

// Peer Role
const (
	// SplitStr linter
	SplitStr = "-"
	// NodeType linter
	NodeType = "pn"
	// PeerUp linter
	PeerUp = "up"
	// PeerDown linter
	PeerDown = "down"
	// PeerNodeType linter
	PeerNodeType = "peer_node"
	// MemberType linter
	MemberType = "member"
)

const (
	// SampleTrackType linter
	SampleTrackType = "sample"
	// RTPTrackType linter
	RTPTrackType = "rtp"
)
