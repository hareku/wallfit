package wallfit

import (
	"encoding/binary"
	"sort"
)

// iccProfileMarker is the APP2 segment identifier for ICC profiles.
var iccProfileMarker = []byte("ICC_PROFILE\x00")

type iccChunk struct {
	seq  int
	data []byte
}

// extractICCFromJPEGBytes extracts the ICC profile bytes from raw JPEG data.
// Returns nil if no ICC profile is found or if the data is not a JPEG.
func extractICCFromJPEGBytes(data []byte) []byte {
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		return nil
	}

	var chunks []iccChunk
	pos := 2

	for pos+3 < len(data) {
		if data[pos] != 0xFF {
			break
		}
		marker := data[pos+1]

		// Skip padding 0xFF bytes.
		if marker == 0xFF {
			pos++
			continue
		}
		// SOI has no length field.
		if marker == 0xD8 {
			pos += 2
			continue
		}
		// EOI or SOS: image data follows, stop scanning.
		if marker == 0xD9 || marker == 0xDA {
			break
		}

		if pos+4 > len(data) {
			break
		}
		// Segment length includes the 2-byte length field itself.
		length := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		if length < 2 {
			break
		}
		segEnd := pos + 2 + length
		if segEnd > len(data) {
			break
		}

		// APP2 (0xE2) with ICC_PROFILE header.
		if marker == 0xE2 && length > 16 {
			payload := data[pos+4 : segEnd]
			if len(payload) >= 14 && string(payload[:12]) == string(iccProfileMarker) {
				seq := int(payload[12])  // 1-based sequence number
				iccData := payload[14:] // skip seq + total bytes
				chunks = append(chunks, iccChunk{seq: seq, data: append([]byte(nil), iccData...)})
			}
		}

		pos = segEnd
	}

	if len(chunks) == 0 {
		return nil
	}

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].seq < chunks[j].seq
	})

	var profile []byte
	for _, c := range chunks {
		profile = append(profile, c.data...)
	}
	return profile
}

// injectICCIntoJPEG inserts the given ICC profile into JPEG bytes immediately
// after the SOI marker. If profile is nil, jpegData is returned unchanged.
func injectICCIntoJPEG(jpegData, profile []byte) []byte {
	if len(profile) == 0 || len(jpegData) < 2 {
		return jpegData
	}

	// Each APP2 segment payload has 14 bytes of overhead:
	//   2-byte length field + 12-byte "ICC_PROFILE\0" + seq byte + total byte
	// Maximum data per segment = 65535 - 14 = 65521 bytes.
	const maxChunkData = 65521
	numChunks := (len(profile) + maxChunkData - 1) / maxChunkData

	var segments []byte
	for i := range numChunks {
		start := i * maxChunkData
		end := min(start+maxChunkData, len(profile))
		chunk := profile[start:end]

		// Segment length = 2 (length field) + 12 (marker) + 1 (seq) + 1 (total) + len(chunk)
		segLen := uint16(2 + 12 + 1 + 1 + len(chunk))
		seg := make([]byte, 2+int(segLen)) // 2 for 0xFF 0xE2
		seg[0] = 0xFF
		seg[1] = 0xE2
		binary.BigEndian.PutUint16(seg[2:4], segLen)
		copy(seg[4:16], iccProfileMarker)
		seg[16] = byte(i + 1)        // 1-based
		seg[17] = byte(numChunks)    // total
		copy(seg[18:], chunk)
		segments = append(segments, seg...)
	}

	result := make([]byte, 0, len(jpegData)+len(segments))
	result = append(result, jpegData[:2]...) // SOI (0xFF 0xD8)
	result = append(result, segments...)
	result = append(result, jpegData[2:]...)
	return result
}
