package fileinfo

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// EBML/Matroska element IDs
const (
	ebmlSegment        = 0x18538067
	ebmlInfo           = 0x1549A966
	ebmlTimestampScale = 0x2AD7B1
	ebmlDuration       = 0x4489
	ebmlTracks         = 0x1654AE6B
	ebmlTrackEntry     = 0xAE
	ebmlTrackType      = 0x83
	ebmlCodecID        = 0x86
	ebmlVideo          = 0xE0
	ebmlPixelWidth     = 0xB0
	ebmlPixelHeight    = 0xBA
	ebmlAudio          = 0xE1
	ebmlDefaultDur     = 0x23E383
)

// parseMKV parses MKV/WebM files natively without ffprobe
func parseMKV(path string) (*FileInfo, error) {
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

	// Skip EBML header
	headerID, headerSize, err := ebmlReadElement(f)
	if err != nil || headerID != 0x1A45DFA3 {
		return nil, fmt.Errorf("not a valid EBML file")
	}
	f.Seek(int64(headerSize), io.SeekCurrent) // skip header data

	// Read Segment
	segID, segSize, err := ebmlReadElement(f)
	if err != nil || segID != ebmlSegment {
		return nil, fmt.Errorf("no Segment element found")
	}

	segStart, _ := f.Seek(0, io.SeekCurrent)
	segEnd := segStart + int64(segSize)

	// Limit parsing to first 10MB of segment to avoid scanning entire file
	parseLimit := segStart + 10*1024*1024
	if parseLimit > segEnd {
		parseLimit = segEnd
	}

	var timestampScale uint64 = 1000000 // default: 1ms
	var durationFloat float64

	for {
		pos, _ := f.Seek(0, io.SeekCurrent)
		if pos >= parseLimit {
			break
		}

		id, size, err := ebmlReadElement(f)
		if err != nil {
			break
		}

		dataStart, _ := f.Seek(0, io.SeekCurrent)

		switch id {
		case ebmlInfo:
			ts, dur := mkvParseInfo(f, size, dataStart)
			if ts > 0 {
				timestampScale = ts
			}
			durationFloat = dur
		case ebmlTracks:
			mkvParseTracks(f, size, dataStart, info)
		default:
			// Skip unknown elements
		}

		f.Seek(dataStart+int64(size), io.SeekStart)
	}

	// Calculate duration
	if durationFloat > 0 && timestampScale > 0 {
		durationSec := (durationFloat * float64(timestampScale)) / 1e9
		info.DurationSeconds = durationSec
		info.Duration = formatDuration(durationSec)
	}

	if info.Codec == "" {
		return nil, fmt.Errorf("no video track found")
	}

	return info, nil
}

// ebmlReadVINT reads a variable-length integer (VINT)
func ebmlReadVINT(r io.Reader) (uint64, int, error) {
	var first [1]byte
	if _, err := io.ReadFull(r, first[:]); err != nil {
		return 0, 0, err
	}

	b := first[0]
	if b == 0 {
		return 0, 0, fmt.Errorf("invalid VINT: leading byte is 0")
	}

	// Count leading zeros to determine length
	var length int
	var mask byte
	for i := 0; i < 8; i++ {
		if b&(0x80>>uint(i)) != 0 {
			length = i + 1
			mask = 0x80 >> uint(i)
			break
		}
	}

	value := uint64(b & ^mask) // remove marker bit

	if length > 1 {
		remaining := make([]byte, length-1)
		if _, err := io.ReadFull(r, remaining); err != nil {
			return 0, 0, err
		}
		for _, rb := range remaining {
			value = (value << 8) | uint64(rb)
		}
	}

	return value, length, nil
}

// ebmlReadElement reads an EBML element ID and data size
func ebmlReadElement(r io.Reader) (id uint64, size uint64, err error) {
	id, _, err = ebmlReadElementID(r)
	if err != nil {
		return 0, 0, err
	}

	size, _, err = ebmlReadVINT(r)
	if err != nil {
		return 0, 0, err
	}

	return id, size, nil
}

// ebmlReadElementID reads an EBML element ID (keeps the marker bit)
func ebmlReadElementID(r io.Reader) (uint64, int, error) {
	var first [1]byte
	if _, err := io.ReadFull(r, first[:]); err != nil {
		return 0, 0, err
	}

	b := first[0]
	if b == 0 {
		return 0, 0, fmt.Errorf("invalid element ID")
	}

	var length int
	for i := 0; i < 8; i++ {
		if b&(0x80>>uint(i)) != 0 {
			length = i + 1
			break
		}
	}

	value := uint64(b) // keep marker bit for ID

	if length > 1 {
		remaining := make([]byte, length-1)
		if _, err := io.ReadFull(r, remaining); err != nil {
			return 0, 0, err
		}
		for _, rb := range remaining {
			value = (value << 8) | uint64(rb)
		}
	}

	return value, length, nil
}

// ebmlReadUint reads an unsigned integer of given size
func ebmlReadUint(r io.Reader, size uint64) (uint64, error) {
	if size > 8 {
		return 0, fmt.Errorf("uint too large: %d bytes", size)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return 0, err
	}
	var val uint64
	for _, b := range data {
		val = (val << 8) | uint64(b)
	}
	return val, nil
}

// ebmlReadFloat reads a float of given size (4 or 8 bytes)
func ebmlReadFloat(r io.Reader, size uint64) (float64, error) {
	if size == 4 {
		var data [4]byte
		if _, err := io.ReadFull(r, data[:]); err != nil {
			return 0, err
		}
		bits := binary.BigEndian.Uint32(data[:])
		return float64(math.Float32frombits(bits)), nil
	} else if size == 8 {
		var data [8]byte
		if _, err := io.ReadFull(r, data[:]); err != nil {
			return 0, err
		}
		bits := binary.BigEndian.Uint64(data[:])
		return math.Float64frombits(bits), nil
	}
	return 0, fmt.Errorf("invalid float size: %d", size)
}

// ebmlReadString reads a string of given size
func ebmlReadString(r io.Reader, size uint64) (string, error) {
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}
	// Trim null bytes
	return strings.TrimRight(string(data), "\x00"), nil
}

// mkvParseInfo parses the Info segment element
func mkvParseInfo(r io.ReadSeeker, size uint64, offset int64) (timestampScale uint64, duration float64) {
	r.Seek(offset, io.SeekStart)
	end := offset + int64(size)

	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			break
		}

		id, sz, err := ebmlReadElement(r)
		if err != nil {
			break
		}

		dataPos, _ := r.Seek(0, io.SeekCurrent)

		switch id {
		case ebmlTimestampScale:
			val, err := ebmlReadUint(r, sz)
			if err == nil {
				timestampScale = val
			}
		case ebmlDuration:
			val, err := ebmlReadFloat(r, sz)
			if err == nil {
				duration = val
			}
		default:
			// skip
		}

		r.Seek(dataPos+int64(sz), io.SeekStart)
	}

	return timestampScale, duration
}

// mkvParseTracks parses the Tracks element
func mkvParseTracks(r io.ReadSeeker, size uint64, offset int64, info *FileInfo) {
	r.Seek(offset, io.SeekStart)
	end := offset + int64(size)

	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			break
		}

		id, sz, err := ebmlReadElement(r)
		if err != nil {
			break
		}

		dataPos, _ := r.Seek(0, io.SeekCurrent)

		if id == ebmlTrackEntry {
			mkvParseTrackEntry(r, sz, dataPos, info)
		}

		r.Seek(dataPos+int64(sz), io.SeekStart)
	}
}

// mkvParseTrackEntry parses a single TrackEntry element
func mkvParseTrackEntry(r io.ReadSeeker, size uint64, offset int64, info *FileInfo) {
	r.Seek(offset, io.SeekStart)
	end := offset + int64(size)

	var trackType uint64
	var codecID string
	var width, height int
	var defaultDuration uint64

	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			break
		}

		id, sz, err := ebmlReadElement(r)
		if err != nil {
			break
		}

		dataPos, _ := r.Seek(0, io.SeekCurrent)

		switch id {
		case ebmlTrackType:
			val, err := ebmlReadUint(r, sz)
			if err == nil {
				trackType = val
			}
		case ebmlCodecID:
			val, err := ebmlReadString(r, sz)
			if err == nil {
				codecID = val
			}
		case ebmlDefaultDur:
			val, err := ebmlReadUint(r, sz)
			if err == nil {
				defaultDuration = val
			}
		case ebmlVideo:
			w, h := mkvParseVideoInfo(r, sz, dataPos)
			width = w
			height = h
		}

		r.Seek(dataPos+int64(sz), io.SeekStart)
	}

	// trackType 1 = video, 2 = audio
	if trackType == 1 && info.Codec == "" {
		info.Codec = mkvCodecName(codecID)
		info.Width = width
		info.Height = height
		if defaultDuration > 0 {
			info.Framerate = math.Round(1e9/float64(defaultDuration)*100) / 100
		}
	}

	if trackType == 2 && info.AudioCodec == "" {
		info.AudioCodec = mkvAudioCodecName(codecID)
	}
}

// mkvParseVideoInfo parses the Video sub-element
func mkvParseVideoInfo(r io.ReadSeeker, size uint64, offset int64) (width, height int) {
	r.Seek(offset, io.SeekStart)
	end := offset + int64(size)

	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			break
		}

		id, sz, err := ebmlReadElement(r)
		if err != nil {
			break
		}

		dataPos, _ := r.Seek(0, io.SeekCurrent)

		switch id {
		case ebmlPixelWidth:
			val, err := ebmlReadUint(r, sz)
			if err == nil {
				width = int(val)
			}
		case ebmlPixelHeight:
			val, err := ebmlReadUint(r, sz)
			if err == nil {
				height = int(val)
			}
		}

		r.Seek(dataPos+int64(sz), io.SeekStart)
	}

	return width, height
}

// mkvCodecName maps Matroska CodecID to human-readable names
func mkvCodecName(codecID string) string {
	switch codecID {
	case "V_MPEG4/ISO/AVC":
		return "h264"
	case "V_MPEGH/ISO/HEVC":
		return "hevc"
	case "V_VP8":
		return "vp8"
	case "V_VP9":
		return "vp9"
	case "V_AV1":
		return "av1"
	case "V_MPEG4/ISO/SP", "V_MPEG4/ISO/ASP", "V_MPEG4/ISO/AP":
		return "mpeg4"
	case "V_MPEG2":
		return "mpeg2video"
	case "V_MPEG1":
		return "mpeg1video"
	default:
		if strings.HasPrefix(codecID, "V_") {
			return strings.ToLower(strings.TrimPrefix(codecID, "V_"))
		}
		return codecID
	}
}

// mkvAudioCodecName maps Matroska audio CodecID to human-readable names
func mkvAudioCodecName(codecID string) string {
	switch codecID {
	case "A_AAC", "A_AAC/MPEG2/LC", "A_AAC/MPEG4/LC":
		return "aac"
	case "A_VORBIS":
		return "vorbis"
	case "A_OPUS":
		return "opus"
	case "A_AC3":
		return "ac3"
	case "A_EAC3":
		return "eac3"
	case "A_DTS":
		return "dts"
	case "A_FLAC":
		return "flac"
	case "A_MPEG/L3":
		return "mp3"
	case "A_MPEG/L2":
		return "mp2"
	case "A_PCM/INT/LIT":
		return "pcm"
	default:
		if strings.HasPrefix(codecID, "A_") {
			return strings.ToLower(strings.TrimPrefix(codecID, "A_"))
		}
		return codecID
	}
}
