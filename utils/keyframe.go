package utils

func IsVP8Keyframe(payload []byte) bool {
	// Check the Payload Descriptor byte (byte 0)
	return payload[0]&0x01 == 0
}

func IsVP9Keyframe(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	pd := new(vp9PayloadDescriptor)
	firstByte := data[0]
	pd.i = (firstByte>>7)&0x01 > 0
	pd.p = (firstByte>>6)&0x01 > 0
	pd.l = (firstByte>>5)&0x01 > 0
	pd.f = (firstByte>>4)&0x01 > 0
	pd.b = (firstByte>>3)&0x01 > 0
	pd.e = (firstByte>>2)&0x01 > 0
	pd.v = (firstByte>>1)&0x01 > 0
	pd.z = (firstByte)&0x01 > 0
	index := 0
	if (firstByte>>7)&0x01 > 0 {
		index++
		if (data[index]>>7)&0x01 > 0 {
			// pd.HasTwoBytesPictureId=true
			index++
		}
	}
	var slIndex byte
	if pd.l {
		index++
		slIndex = (data[index] >> 1) & 0x07
	}
	if !pd.p && pd.b && slIndex == 0 {
		pd.isKeyFrame = true
	}

	return pd.isKeyFrame
}

func IsH264Keyframe(payload []byte) bool {
	nalType := payload[0] & 0x1F
	return nalType == 5 || nalType == 7 || nalType == 8
}

type vp9PayloadDescriptor struct {
	i bool
	p bool
	l bool
	f bool
	b bool
	e bool
	v bool
	z bool

	twoBytePictureID bool
	pictureID        uint16 // up to 2 byte

	isKeyFrame bool
}
