package backupmgr

import (
	"bytes"
	"io"
	"strconv"
)

// The game emits control-character references such as &#xB; in custom names.
// Normalize those for Go's XML decoder only. Archive bytes and ZIP CRCs remain
// untouched, and malformed tags/entities still fail strict structural validation.
type saveXMLReader struct {
	reader  io.Reader
	buffer  [32 * 1024]byte
	pending []byte
	tail    []byte
	readErr error
}

func (r *saveXMLReader) Read(output []byte) (int, error) {
	if len(output) == 0 {
		return 0, nil
	}
	for len(r.pending) == 0 && r.readErr == nil {
		prefix := copy(r.buffer[:], r.tail)
		n, err := r.reader.Read(r.buffer[prefix:])
		r.readErr = err
		data := r.buffer[:prefix+n]
		end := len(data)
		r.tail = nil
		for i := 0; i < len(data); i++ {
			if illegalXMLControl(uint64(data[i])) {
				data[i] = ' '
				continue
			}
			if data[i] != '&' {
				continue
			}
			// Numeric XML references fit comfortably in 16 bytes. Carry a possible
			// partial reference across reads, including one-byte underlying readers.
			length := min(16, len(data)-i)
			semi := bytes.IndexByte(data[i:i+length], ';')
			if semi < 0 {
				if length < 16 && err == nil {
					end = i
					r.tail = append([]byte(nil), data[i:]...)
					break
				}
				continue
			}
			entity := data[i+1 : i+semi]
			if len(entity) < 2 || entity[0] != '#' {
				continue
			}
			digits, base := entity[1:], 10
			if digits[0] == 'x' {
				digits = digits[1:]
				base = 16
			}
			value, parseErr := strconv.ParseUint(string(digits), base, 32)
			if parseErr == nil && illegalXMLControl(value) {
				for j := i; j <= i+semi; j++ {
					data[j] = ' '
				}
			}
			i += semi
		}
		r.pending = data[:end]
	}
	if len(r.pending) > 0 {
		n := copy(output, r.pending)
		r.pending = r.pending[n:]
		return n, nil
	}
	return 0, r.readErr
}

func illegalXMLControl(value uint64) bool {
	return value < 0x20 && value != '\t' && value != '\n' && value != '\r'
}
