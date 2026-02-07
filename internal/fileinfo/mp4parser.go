package fileinfo

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

// parseMP4 parses MP4/MOV/M4V files natively without ffprobe
func parseMP4(path string) (*FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	info := &FileInfo{
		Path:     path,
		Name:     filepath.Base(path),
		FileSize: stat.Size(),
	}

	// Parse top-level boxes to find moov
	err = mp4IterateBoxes(f, stat.Size(), 0, func(boxType string, dataSize int64, dataOffset int64) error {
		if boxType == "moov" {
			return mp4ParseMoov(f, dataSize, dataOffset, info)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if info.Codec == "" {
		return nil, fmt.Errorf("no video track found")
	}

	return info, nil
}

// mp4IterateBoxes iterates boxes within a container
func mp4IterateBoxes(r io.ReadSeeker, containerSize int64, containerOffset int64, handler func(string, int64, int64) error) error {
	endPos := containerOffset + containerSize

	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos+8 > endPos {
			break
		}

		var header [8]byte
		if _, err := io.ReadFull(r, header[:]); err != nil {
			break
		}

		size := int64(binary.BigEndian.Uint32(header[:4]))
		boxType := string(header[4:8])
		headerSize := int64(8)

		if size == 1 {
			var ext [8]byte
			if _, err := io.ReadFull(r, ext[:]); err != nil {
				break
			}
			size = int64(binary.BigEndian.Uint64(ext[:]))
			headerSize = 16
		} else if size == 0 {
			size = endPos - pos
		}

		if size < headerSize || size > endPos-pos {
			break
		}

		dataSize := size - headerSize
		dataOffset, _ := r.Seek(0, io.SeekCurrent)

		if err := handler(boxType, dataSize, dataOffset); err != nil {
			return err
		}

		// Skip to next box
		if _, err := r.Seek(pos+size, io.SeekStart); err != nil {
			break
		}
	}

	return nil
}

// mp4ParseMoov parses the moov container
func mp4ParseMoov(r io.ReadSeeker, size int64, offset int64, info *FileInfo) error {
	r.Seek(offset, io.SeekStart)

	var movieTimescale uint32
	var movieDuration uint64

	return mp4IterateBoxes(r, size, offset, func(boxType string, dataSize int64, dataOffset int64) error {
		switch boxType {
		case "mvhd":
			ts, dur, err := mp4ParseMvhd(r, dataSize, dataOffset)
			if err == nil {
				movieTimescale = ts
				movieDuration = dur
				if movieTimescale > 0 {
					durationSec := float64(movieDuration) / float64(movieTimescale)
					info.DurationSeconds = durationSec
					info.Duration = formatDuration(durationSec)
				}
			}
		case "trak":
			mp4ParseTrak(r, dataSize, dataOffset, info, movieTimescale)
		}
		return nil
	})
}

// mp4ParseMvhd parses the movie header box
func mp4ParseMvhd(r io.ReadSeeker, size int64, offset int64) (timescale uint32, duration uint64, err error) {
	r.Seek(offset, io.SeekStart)

	var version [1]byte
	if _, err := io.ReadFull(r, version[:]); err != nil {
		return 0, 0, err
	}

	if version[0] == 0 {
		// Version 0: 4-byte fields
		r.Seek(offset+4, io.SeekStart) // skip version(1) + flags(3)
		var data [16]byte              // create_time(4) + modify_time(4) + timescale(4) + duration(4)
		if _, err := io.ReadFull(r, data[:]); err != nil {
			return 0, 0, err
		}
		timescale = binary.BigEndian.Uint32(data[8:12])
		duration = uint64(binary.BigEndian.Uint32(data[12:16]))
	} else {
		// Version 1: 8-byte fields
		r.Seek(offset+4, io.SeekStart) // skip version(1) + flags(3)
		var data [28]byte              // create_time(8) + modify_time(8) + timescale(4) + duration(8)
		if _, err := io.ReadFull(r, data[:]); err != nil {
			return 0, 0, err
		}
		timescale = binary.BigEndian.Uint32(data[16:20])
		duration = binary.BigEndian.Uint64(data[20:28])
	}

	return timescale, duration, nil
}

// mp4ParseTrak parses a track container
func mp4ParseTrak(r io.ReadSeeker, size int64, offset int64, info *FileInfo, movieTimescale uint32) {
	r.Seek(offset, io.SeekStart)

	var trackWidth, trackHeight int
	var handlerType string
	var codec string
	var mediaTimescale uint32
	var sampleCount uint32
	var mediaDuration uint64

	mp4IterateBoxes(r, size, offset, func(boxType string, dataSize int64, dataOffset int64) error {
		switch boxType {
		case "tkhd":
			w, h := mp4ParseTkhd(r, dataSize, dataOffset)
			trackWidth = w
			trackHeight = h
		case "mdia":
			r.Seek(dataOffset, io.SeekStart)
			mp4IterateBoxes(r, dataSize, dataOffset, func(boxType string, dataSize int64, dataOffset int64) error {
				switch boxType {
				case "mdhd":
					ts, dur := mp4ParseMdhd(r, dataSize, dataOffset)
					mediaTimescale = ts
					mediaDuration = dur
				case "hdlr":
					handlerType = mp4ParseHdlr(r, dataSize, dataOffset)
				case "minf":
					r.Seek(dataOffset, io.SeekStart)
					mp4IterateBoxes(r, dataSize, dataOffset, func(boxType string, dataSize int64, dataOffset int64) error {
						if boxType == "stbl" {
							r.Seek(dataOffset, io.SeekStart)
							mp4IterateBoxes(r, dataSize, dataOffset, func(boxType string, dataSize int64, dataOffset int64) error {
								switch boxType {
								case "stsd":
									codec = mp4ParseStsd(r, dataSize, dataOffset)
								case "stts":
									sampleCount = mp4ParseStts(r, dataSize, dataOffset)
								}
								return nil
							})
						}
						return nil
					})
				}
				return nil
			})
		}
		return nil
	})

	if handlerType == "vide" && info.Codec == "" {
		info.Codec = mp4CodecName(codec)
		info.Width = trackWidth
		info.Height = trackHeight

		// Calculate framerate from media timescale and sample count
		if mediaTimescale > 0 && sampleCount > 0 && mediaDuration > 0 {
			durationSec := float64(mediaDuration) / float64(mediaTimescale)
			if durationSec > 0 {
				fps := float64(sampleCount) / durationSec
				info.Framerate = math.Round(fps*100) / 100
			}
		}
	}

	if handlerType == "soun" && info.AudioCodec == "" {
		info.AudioCodec = mp4AudioCodecName(codec)
	}
}

// mp4ParseTkhd parses track header for width/height
func mp4ParseTkhd(r io.ReadSeeker, size int64, offset int64) (width, height int) {
	r.Seek(offset, io.SeekStart)

	var version [1]byte
	io.ReadFull(r, version[:])

	if version[0] == 0 {
		// Version 0: skip to width/height at offset 76 from start of box data
		r.Seek(offset+76, io.SeekStart)
	} else {
		// Version 1: skip to width/height at offset 88
		r.Seek(offset+88, io.SeekStart)
	}

	var dims [8]byte // width(4) + height(4) in 16.16 fixed-point
	if _, err := io.ReadFull(r, dims[:]); err != nil {
		return 0, 0
	}

	width = int(binary.BigEndian.Uint32(dims[0:4]) >> 16)
	height = int(binary.BigEndian.Uint32(dims[4:8]) >> 16)
	return width, height
}

// mp4ParseMdhd parses media header for timescale and duration
func mp4ParseMdhd(r io.ReadSeeker, size int64, offset int64) (timescale uint32, duration uint64) {
	r.Seek(offset, io.SeekStart)

	var version [1]byte
	io.ReadFull(r, version[:])

	if version[0] == 0 {
		r.Seek(offset+12, io.SeekStart) // skip version(1)+flags(3)+create(4)+modify(4)
		var tsData [8]byte              // timescale(4) + duration(4)
		io.ReadFull(r, tsData[:])
		timescale = binary.BigEndian.Uint32(tsData[0:4])
		duration = uint64(binary.BigEndian.Uint32(tsData[4:8]))
	} else {
		r.Seek(offset+20, io.SeekStart) // skip version(1)+flags(3)+create(8)+modify(8)
		var tsData [12]byte
		io.ReadFull(r, tsData[:])
		timescale = binary.BigEndian.Uint32(tsData[0:4])
		duration = binary.BigEndian.Uint64(tsData[4:12])
	}

	return timescale, duration
}

// mp4ParseHdlr parses handler box to get track type
func mp4ParseHdlr(r io.ReadSeeker, size int64, offset int64) string {
	r.Seek(offset+8, io.SeekStart) // skip version(1)+flags(3)+pre_defined(4)

	var handlerType [4]byte
	if _, err := io.ReadFull(r, handlerType[:]); err != nil {
		return ""
	}

	return string(handlerType[:])
}

// mp4ParseStsd parses sample description to get codec
func mp4ParseStsd(r io.ReadSeeker, size int64, offset int64) string {
	r.Seek(offset+8, io.SeekStart) // skip version(1)+flags(3)+entry_count(4)

	// Read first sample entry's codec type (skip size(4), read type(4))
	var entry [8]byte
	if _, err := io.ReadFull(r, entry[:]); err != nil {
		return ""
	}

	return string(entry[4:8])
}

// mp4ParseStts parses time-to-sample table to get total sample count
func mp4ParseStts(r io.ReadSeeker, size int64, offset int64) uint32 {
	r.Seek(offset+4, io.SeekStart) // skip version(1)+flags(3)

	var entryCount [4]byte
	if _, err := io.ReadFull(r, entryCount[:]); err != nil {
		return 0
	}

	count := binary.BigEndian.Uint32(entryCount[:])
	var totalSamples uint32

	for i := uint32(0); i < count; i++ {
		var entry [8]byte // sample_count(4) + sample_delta(4)
		if _, err := io.ReadFull(r, entry[:]); err != nil {
			break
		}
		totalSamples += binary.BigEndian.Uint32(entry[0:4])
	}

	return totalSamples
}

// mp4CodecName maps MP4 codec FourCC to human-readable names
func mp4CodecName(fourcc string) string {
	switch fourcc {
	case "avc1", "avc3":
		return "h264"
	case "hev1", "hvc1":
		return "hevc"
	case "vp08":
		return "vp8"
	case "vp09":
		return "vp9"
	case "av01":
		return "av1"
	case "mp4v":
		return "mpeg4"
	default:
		if fourcc != "" {
			return fourcc
		}
		return ""
	}
}

// mp4AudioCodecName maps MP4 audio codec FourCC to human-readable names
func mp4AudioCodecName(fourcc string) string {
	switch fourcc {
	case "mp4a":
		return "aac"
	case "ac-3":
		return "ac3"
	case "ec-3":
		return "eac3"
	case "Opus":
		return "opus"
	case "fLaC":
		return "flac"
	case "alac":
		return "alac"
	default:
		if fourcc != "" {
			return fourcc
		}
		return ""
	}
}
