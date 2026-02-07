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

// parseAVI parses AVI files natively without ffprobe
func parseAVI(path string) (*FileInfo, error) {
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

	// Read RIFF header
	var riffHeader [12]byte
	if _, err := io.ReadFull(f, riffHeader[:]); err != nil {
		return nil, fmt.Errorf("not a valid RIFF file")
	}

	if string(riffHeader[0:4]) != "RIFF" || string(riffHeader[8:12]) != "AVI " {
		return nil, fmt.Errorf("not an AVI file")
	}

	riffSize := int64(binary.LittleEndian.Uint32(riffHeader[4:8]))

	// Parse chunks
	parseLimit := int64(12) + riffSize
	if parseLimit > stat.Size() {
		parseLimit = stat.Size()
	}

	// Limit parsing to first 1MB for speed
	if parseLimit > 1*1024*1024 {
		parseLimit = 1 * 1024 * 1024
	}

	err = aviIterateChunks(f, 12, parseLimit, info)
	if err != nil {
		return nil, err
	}

	if info.Codec == "" {
		return nil, fmt.Errorf("no video stream found")
	}

	return info, nil
}

// aviIterateChunks iterates over RIFF chunks
func aviIterateChunks(r io.ReadSeeker, startPos, endPos int64, info *FileInfo) error {
	r.Seek(startPos, io.SeekStart)

	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos+8 > endPos {
			break
		}

		var chunkHeader [8]byte
		if _, err := io.ReadFull(r, chunkHeader[:]); err != nil {
			break
		}

		chunkID := string(chunkHeader[0:4])
		chunkSize := int64(binary.LittleEndian.Uint32(chunkHeader[4:8]))

		dataPos, _ := r.Seek(0, io.SeekCurrent)

		switch chunkID {
		case "LIST":
			// Read list type
			var listType [4]byte
			if _, err := io.ReadFull(r, listType[:]); err != nil {
				break
			}
			lt := string(listType[:])

			listDataPos, _ := r.Seek(0, io.SeekCurrent)

			switch lt {
			case "hdrl":
				aviIterateChunks(r, listDataPos, dataPos+chunkSize, info)
			case "strl":
				aviParseStreamList(r, listDataPos, dataPos+chunkSize, info)
			}
		case "avih":
			aviParseMainHeader(r, chunkSize, dataPos, info)
		}

		// Align to word boundary
		nextPos := dataPos + chunkSize
		if nextPos%2 != 0 {
			nextPos++
		}
		r.Seek(nextPos, io.SeekStart)
	}

	return nil
}

// aviParseMainHeader parses the main AVI header (avih chunk)
func aviParseMainHeader(r io.ReadSeeker, size int64, offset int64, info *FileInfo) {
	r.Seek(offset, io.SeekStart)

	if size < 56 {
		return
	}

	var data [56]byte
	if _, err := io.ReadFull(r, data[:]); err != nil {
		return
	}

	microSecPerFrame := binary.LittleEndian.Uint32(data[0:4])
	totalFrames := binary.LittleEndian.Uint32(data[24:28])
	width := binary.LittleEndian.Uint32(data[32:36])
	height := binary.LittleEndian.Uint32(data[36:40])

	if info.Width == 0 && width > 0 {
		info.Width = int(width)
	}
	if info.Height == 0 && height > 0 {
		info.Height = int(height)
	}

	// Calculate duration from total frames and microseconds per frame
	if microSecPerFrame > 0 && totalFrames > 0 {
		fps := 1e6 / float64(microSecPerFrame)
		durationSec := float64(totalFrames) / fps
		info.DurationSeconds = durationSec
		info.Duration = formatDuration(durationSec)

		if info.Framerate == 0 {
			info.Framerate = math.Round(fps*100) / 100
		}
	}
}

// aviParseStreamList parses a stream list (strl) to extract stream header and format
func aviParseStreamList(r io.ReadSeeker, startPos, endPos int64, info *FileInfo) {
	r.Seek(startPos, io.SeekStart)

	var streamType string
	var codecFourCC string

	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos+8 > endPos {
			break
		}

		var chunkHeader [8]byte
		if _, err := io.ReadFull(r, chunkHeader[:]); err != nil {
			break
		}

		chunkID := string(chunkHeader[0:4])
		chunkSize := int64(binary.LittleEndian.Uint32(chunkHeader[4:8]))
		dataPos, _ := r.Seek(0, io.SeekCurrent)

		switch chunkID {
		case "strh":
			if chunkSize >= 8 {
				var strh [56]byte
				n := chunkSize
				if n > 56 {
					n = 56
				}
				if _, err := io.ReadFull(r, strh[:n]); err == nil {
					streamType = strings.TrimRight(string(strh[0:4]), "\x00")
					codecFourCC = strings.TrimRight(string(strh[4:8]), "\x00")

					// Parse framerate from dwScale and dwRate
					if n >= 28 {
						dwScale := binary.LittleEndian.Uint32(strh[20:24])
						dwRate := binary.LittleEndian.Uint32(strh[24:28])
						if dwScale > 0 && dwRate > 0 {
							fps := float64(dwRate) / float64(dwScale)
							if streamType == "vids" && info.Framerate == 0 {
								info.Framerate = math.Round(fps*100) / 100
							}
						}
					}

					// Parse duration from dwLength
					if n >= 36 {
						dwScale := binary.LittleEndian.Uint32(strh[20:24])
						dwRate := binary.LittleEndian.Uint32(strh[24:28])
						dwLength := binary.LittleEndian.Uint32(strh[32:36])
						if dwScale > 0 && dwRate > 0 && dwLength > 0 {
							durationSec := float64(dwLength) * float64(dwScale) / float64(dwRate)
							if streamType == "vids" && durationSec > 0 {
								info.DurationSeconds = durationSec
								info.Duration = formatDuration(durationSec)
							}
						}
					}
				}
			}
		case "strf":
			if streamType == "vids" && chunkSize >= 40 {
				// BITMAPINFOHEADER
				var bih [40]byte
				if _, err := io.ReadFull(r, bih[:]); err == nil {
					w := int(binary.LittleEndian.Uint32(bih[4:8]))
					h := int(binary.LittleEndian.Uint32(bih[8:12]))
					if h < 0 {
						h = -h // height can be negative for top-down
					}
					if w > 0 {
						info.Width = w
					}
					if h > 0 {
						info.Height = h
					}
				}
			}
		}

		// Align to word boundary
		nextPos := dataPos + chunkSize
		if nextPos%2 != 0 {
			nextPos++
		}
		r.Seek(nextPos, io.SeekStart)
	}

	// Set codec info
	if streamType == "vids" && info.Codec == "" {
		info.Codec = aviCodecName(codecFourCC)
	}
	if streamType == "auds" && info.AudioCodec == "" {
		info.AudioCodec = aviAudioCodecName(codecFourCC)
	}
}

// aviCodecName maps AVI video FourCC to human-readable names
func aviCodecName(fourcc string) string {
	upper := strings.ToUpper(fourcc)
	switch upper {
	case "H264", "X264", "AVC1":
		return "h264"
	case "HEVC", "H265", "X265", "HVC1":
		return "hevc"
	case "DIVX", "DX50", "XVID", "FMP4", "MP4V":
		return "mpeg4"
	case "MJPG", "MJPEG":
		return "mjpeg"
	case "VP80":
		return "vp8"
	case "VP90":
		return "vp9"
	case "AV01":
		return "av1"
	case "WMV1":
		return "wmv1"
	case "WMV2":
		return "wmv2"
	case "WMV3":
		return "wmv3"
	case "MSVC", "CRAM":
		return "msvideo1"
	default:
		if fourcc != "" {
			return strings.ToLower(fourcc)
		}
		return ""
	}
}

// aviAudioCodecName maps AVI audio handler/tag to human-readable names
func aviAudioCodecName(handler string) string {
	upper := strings.ToUpper(handler)
	switch upper {
	case "\x00\x00\x00\x00", "", "\x01\x00":
		return "pcm"
	case "\x55\x00":
		return "mp3"
	case "\xFF\x00":
		return "aac"
	default:
		if handler != "" {
			return strings.ToLower(strings.TrimRight(handler, "\x00"))
		}
		return ""
	}
}
