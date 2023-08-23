package utils

func IsVP8Keyframe(payload []byte) bool {
	// Check the Payload Descriptor byte (byte 0)
	return payload[0]&0x01 == 0
}

func IsVP9Keyframe(payload []byte) bool {
	// Check the S (Start of Frame) bit in the Payload Descriptor byte (byte 0)
	return payload[0]&0x01 == 0
}

func IsH264Keyframe(payload []byte) bool {
	nalType := payload[0] & 0x1F
	return nalType == 5 || nalType == 7 || nalType == 8
}
