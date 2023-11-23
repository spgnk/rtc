package utils

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"log"
	"os"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
	"github.com/xlab/libvpx-go/vpx"
)

type VDecoder struct {
	enabled bool

	src   <-chan []byte
	ctx   *vpx.CodecCtx
	iface *vpx.CodecIface
}

type VCodec string

const (
	CodecVP8 VCodec = MimeTypeVP8
	CodecVP9 VCodec = MimeTypeVP9
)

func NewVDecoder(codec VCodec, src <-chan []byte) *VDecoder {
	dec := &VDecoder{
		src: src,
		ctx: vpx.NewCodecCtx(),
	}
	switch codec {
	case CodecVP8:
		dec.iface = vpx.DecoderIfaceVP8()
	case CodecVP9:
		dec.iface = vpx.DecoderIfaceVP9()
	default: // others are currently disabled
		log.Println("[WARN] unsupported VPX codec:", codec)
		return dec
	}
	err := vpx.Error(vpx.CodecDecInitVer(dec.ctx, dec.iface, nil, 0, vpx.DecoderABIVersion))
	if err != nil {
		log.Println("[WARN]", err)
		return dec
	}
	dec.enabled = true
	return dec
}

func (v *VDecoder) Save(savePath string) { //, out chan<- Frame
	//  defer close(out)
	i := 0

	sampleBuilder := samplebuilder.New(20000, &codecs.VP8Packet{}, 90000)

	for data := range v.src {
		r := &rtp.Packet{}
		if err := r.Unmarshal(data); err != nil {
			log.Println("[WARN]", err)
			continue
		}
		sampleBuilder.Push(r)
		// Use SampleBuilder to generate full picture from many RTP Packets
		sample := sampleBuilder.Pop()
		if sample == nil {
			continue
		}

		if !v.enabled {
			continue
		}

		dataSize := uint32(len(sample.Data))

		err := vpx.Error(vpx.CodecDecode(v.ctx, string(sample.Data), dataSize, nil, 0))
		if err != nil {
			log.Println("[WARN]", err)
			continue
		}

		var iter vpx.CodecIter
		img := vpx.CodecGetFrame(v.ctx, &iter)
		if img != nil {
			img.Deref()

			// out <- Frame{
			//  RGBA:     img.ImageRGBA(),
			//  Timecode: time.Duration(pkt.Timestamp),
			// }

			i++

			buffer := new(bytes.Buffer)
			if err = jpeg.Encode(buffer, img.ImageYCbCr(), nil); err != nil {
				//  panic(err)
				fmt.Printf("jpeg Encode Error: %s\r\n", err)
			}

			fo, err := os.Create(fmt.Sprintf("%s%d%s", savePath, i, ".jpg"))

			if err != nil {
				fmt.Printf("image create Error: %s\r\n", err)
				//panic(err)
			}
			// close fo on exit and check for its returned error
			defer func() {
				if err := fo.Close(); err != nil {
					panic(err)
				}
			}()

			if _, err := fo.Write(buffer.Bytes()); err != nil {
				fmt.Printf("image write Error: %s\r\n", err)
				//panic(err)
			}

			fo.Close()
		}
	}
}
