package utils

import (
	"os"
	"strconv"

	"github.com/pion/webrtc/v3"
)

var (
	// DEBUG to logging debug
	DEBUG = os.Getenv("DEBUG")
	// TIMEOUT linter
	TIMEOUT = os.Getenv("TIMEOUT")
)

// NewSDPType format sdp type between pion and javascript
func NewSDPType(raw string) webrtc.SDPType {
	switch raw {
	case "offer":
		return webrtc.SDPTypeOffer
	case "answer":
		return webrtc.SDPTypeAnswer
	default:
		return webrtc.SDPType(webrtc.Unknown)
	}
}

// GetRepeerGaptime linter
func GetRepeerGaptime() int {
	i := 5
	if interval := os.Getenv("REPEER_GAP_TIME"); interval != "" {
		j, err := strconv.Atoi(interval)
		if err == nil {
			i = j
		}
	}
	return i
}

const defaultInterval = 15

// GetInterval linter
func GetInterval() int {
	i := defaultInterval
	if interval := os.Getenv("SYNC_INTERVAL"); interval != "" {
		j, err := strconv.Atoi(interval)
		if err == nil {
			i = j
		}
	}
	return i
}
