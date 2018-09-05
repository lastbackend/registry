//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2018] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package cleaner

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// These parameters should equals to https://github.com/docker/docker/blob/master/pkg/stdcopy/stdcopy.go
	stdWriterPrefixLen = 8 // size of prefix header
	stdWriterSizeIndex = 4 // size byte index in header

	// Default reader configuration
	defaultBufferLength = 1024 * 2
	defaultDataLength   = 1024 * 64
)

type reader struct {
	reader io.Reader

	size   uint32
	cursor uint32

	buffer []byte
	prefix []byte

	cleared bool
}

// NewReader returns a reader that strips off the message headers from raw docker logs stream
func NewReader(r io.Reader) io.Reader {
	return &reader{
		reader: r,
		prefix: make([]byte, stdWriterPrefixLen),
		buffer: make([]byte, defaultBufferLength)}
}

func (r *reader) Read(p []byte) (int, error) {

	if !r.cleared {
		// parse and store the message
		if err := r.clear(); err != nil {
			return 0, err
		}
		r.cleared = true
	}

	n, err := r.read(p)
	if err == io.EOF {
		err = nil // continue next chunk (clear() handles the EOF from r.reader)
		r.cleared = false
	}

	return n, nil
}

func (r *reader) read(p []byte) (int, error) {
	if r.size <= r.cursor {
		r.cleared = false
		return 0, io.EOF
	}
	n := copy(p, r.buffer[r.cursor:r.size])
	r.cursor += uint32(n)
	return n, nil
}

func (r *reader) clear() error {

	n, err := io.ReadFull(r.reader, r.prefix)
	if err != nil {
		switch err {
		case context.Canceled:
			return err
		case io.EOF:
			return err // end of the underlying logs stream
		case io.ErrUnexpectedEOF:
			return fmt.Errorf("defective prefix read of %d bytes", n)
		default:
			return fmt.Errorf("reading prefix err: %v", err)
		}
	}

	if r.prefix[0] != 0x1 && r.prefix[0] != 0x2 {
		return fmt.Errorf("unexpected stream byte: %#x", r.prefix[0])
	}

	size := binary.BigEndian.Uint32(r.prefix[stdWriterSizeIndex : stdWriterSizeIndex+4])
	// prevent reading garbage
	if size > defaultDataLength {
		return fmt.Errorf("exceeded the data limit (%d/%d) bytes", size, defaultDataLength)
	}

	if int(size) > len(r.buffer) {
		// increase the buffer if necessary
		r.buffer = make([]byte, size)
	}

	m, err := io.ReadFull(r.reader, r.buffer[:int(size)])
	if err != nil {
		switch err {
		case context.Canceled:
			return err
		case io.EOF, io.ErrUnexpectedEOF:
			return fmt.Errorf("read message %d out of %d bytes err: %v", m, size, err)
		default:
			return fmt.Errorf("read message err: %v", err)
		}
	}

	// reset cursor for next chunk
	r.size = size
	r.cursor = uint32(0)

	return nil
}
