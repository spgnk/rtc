package utils

import (
	"strings"

	"github.com/lamhai1401/sdpParser/sdp"
)

// SDPParser linter
type SDPParser struct {
	parser sdp.SessionDescription // handld sdp parse
}

// NewSDPParser linter
func NewSDPParser(data *string) (*SDPParser, error) {
	parser := &SDPParser{}

	temp := sdp.SessionDescription{}
	err := temp.Unmarshal([]byte(*data))
	if err != nil {
		return nil, err
	}
	parser.parser = temp
	return parser, nil
}

// GetPayLoadType input a sdp string return payload of input codec
func (s *SDPParser) GetPayLoadType(codec *string) (uint8, error) {
	vp := sdp.Codec{
		Name: strings.ToUpper(*codec),
	}
	payLoadType, err := s.parser.GetPayloadTypeForCodec(vp)
	if err != nil {
		return 0, err
	}
	return payLoadType, nil
}

// GetArrPayLoadType return a list of input sdp payload input sdp data
func (s *SDPParser) GetArrPayLoadType(codec *string) ([]uint8, error) {
	vp := sdp.Codec{
		Name: strings.ToUpper(*codec),
	}
	result, err := s.parser.GetPayloadTypeForCodecs(vp)
	if err != nil {
		return result, err
	}

	return result, nil
}
