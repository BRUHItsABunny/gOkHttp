package gokhttp_download

import (
	"bytes"
	"fmt"
	"github.com/yapingcat/gomedia/go-mpeg2"
	"os"
)

type StreamMuxer struct {
	F       *os.File
	Muxer   *mpeg2.TSMuxer
	Demuxer *mpeg2.TSDemuxer
	Streams map[mpeg2.TS_STREAM_TYPE]uint16
}

func NewStreamMuxer(f *os.File) *StreamMuxer {
	result := &StreamMuxer{
		F:       f,
		Muxer:   mpeg2.NewTSMuxer(),
		Demuxer: mpeg2.NewTSDemuxer(),
		Streams: map[mpeg2.TS_STREAM_TYPE]uint16{},
	}

	result.Muxer.OnPacket = result.Mux
	result.Demuxer.OnFrame = result.Demux

	return result
}

func (m *StreamMuxer) AddStreams(buf *bytes.Buffer) error {
	tsPremuxer := mpeg2.NewTSDemuxer()
	videoAdded := false
	tsPremuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
		streamId, ok := m.Streams[cid]
		if !ok {
			switch cid {
			case mpeg2.TS_STREAM_AAC:
				fallthrough
			case mpeg2.TS_STREAM_AUDIO_MPEG1:
				fallthrough
			case mpeg2.TS_STREAM_AUDIO_MPEG2:
				if videoAdded {
					streamId = m.Muxer.AddStream(cid)
					m.Streams[cid] = streamId
				}
				break
			default:
				streamId = m.Muxer.AddStream(cid)
				m.Streams[cid] = streamId
				videoAdded = true
			}
		}
	}
	err := tsPremuxer.Input(buf)
	if err != nil {
		return fmt.Errorf("tsPremuxer.Input: %w", err)
	}
	return nil
}

func (m *StreamMuxer) Mux(pkg []byte) {
	_, err := m.F.Write(pkg)
	if err != nil {
		panic(fmt.Errorf("result.F.Write: %w", err))
	}
}

func (m *StreamMuxer) Demux(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
	streamId, ok := m.Streams[cid]
	if !ok {
		return
	}

	/*
		switch cid {
		case mpeg2.TS_STREAM_H264:
			codec.SplitFrameWithStartCode(frame, func(nalu []byte) bool {
				err := m.Muxer.Write(streamId, nalu, pts, dts)
				if err != nil {
					panic(fmt.Errorf("[h264] result.Muxer.Write: %w", err))
				}
				return true
			})
			break
		case mpeg2.TS_STREAM_H265:
			codec.SplitFrameWithStartCode(frame, func(nalu []byte) bool {
				err := m.Muxer.Write(streamId, nalu, pts, dts)
				if err != nil {
					panic(fmt.Errorf("[h265] result.Muxer.Write: %w", err))
				}
				return true
			})
			break
		case mpeg2.TS_STREAM_AAC:
			codec.SplitAACFrame(frame, func(aac []byte) {
				err := m.Muxer.Write(streamId, aac, pts, dts)
				if err != nil {
					panic(fmt.Errorf("[aac] result.Muxer.Write: %w", err))
				}
			})
			break
		case mpeg2.TS_STREAM_AUDIO_MPEG1:
			fallthrough
		case mpeg2.TS_STREAM_AUDIO_MPEG2:
			err := codec.SplitMp3Frames(frame, func(head *codec.MP3FrameHead, mp3 []byte) {
				err := m.Muxer.Write(streamId, mp3, pts, dts)
				if err != nil {
					panic(fmt.Errorf("[mp3] result.Muxer.Write: %w", err))
				}
			})
			if err != nil {
				panic(fmt.Errorf("codec.SplitMp3Frames: %w", err))
			}
		}
	*/

	err := m.Muxer.Write(streamId, frame, pts, dts)
	if err != nil {
		panic(fmt.Errorf("result.Muxer.Write: %w", err))
	}
}
