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

package logger

import (
	"context"
	"io"
	"os"

	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/stream"
)

const (
	DEFAULT_TEMP_BUFFER_SIZE = 1024
	DEFAULT_BUFFER_SIZE      = 5e+6 //  5e+6 = 5MB
)

type Logger struct {
	ctx    context.Context
	size   int // Buffer fullness
	count  int // Number of bytes written since last call to Count()
	buffer []byte

	chunk chan []byte
	done  chan bool

	stdout bool

	streams map[*stream.Stream]bool
	writers map[io.WriteCloser]bool
}

func NewLogger(ctx context.Context) *Logger {

	l := &Logger{
		ctx:     ctx,
		buffer:  make([]byte, 0, DEFAULT_BUFFER_SIZE),
		chunk:   make(chan []byte),
		done:    make(chan bool, 1),
		streams: make(map[*stream.Stream]bool),
		writers: make(map[io.WriteCloser]bool),
	}

	// Loop processing: send chunk to writers
	go func() {
		for {
			select {
			case chunk := <-l.chunk:

				// Send to all streams
				for s := range l.streams {
					_, err := func(p []byte) (n int, err error) {
						n, err = s.Write(p)
						if err == nil {
							s.Flush()
						}
						return n, err
					}(chunk)

					if err != nil {
						log.Errorf("Error written to stream %s", err)
						return
					}
				}

				// write to all writers
				for w := range l.writers {
					_, err := func(p []byte) (n int, err error) {
						n, err = w.Write(p)
						return n, err
					}(chunk)

					if err != nil {
						log.Errorf("Error written to stream %s", err)
						return
					}
				}
			}
		}
	}()

	return l
}

// Read bytes for buffer from reader
// stdout - write logs to stdout
// filePath - write logs to file
func (l *Logger) Run(reader io.ReadCloser, stdout bool) error {

	if reader == nil {
		return nil
	}

	for {

		buffer := make([]byte, DEFAULT_TEMP_BUFFER_SIZE)

		readBytes, err := reader.Read(buffer)

		if err != nil && err != io.EOF {
			return err
		}
		if readBytes == 0 {
			l.done <- true
			break
		}

		if stdout {
			os.Stdout.Write(buffer[:readBytes])
		}

		l.size += readBytes
		l.count = readBytes

		l.buffer = append(l.buffer, buffer[:readBytes]...)

		l.chunk <- buffer[:readBytes]
	}

	return nil
}

// Adds stream to streams to further redirect data from the read reader
func (l *Logger) Stream(stream *stream.Stream) error {

	var done = make(chan bool, 1)

	go func() {
		stream.Done()
		done <- true
	}()

	_, err := func(p []byte) (n int, err error) {
		n, err = stream.Write(p)
		if err == nil {
			stream.Flush()
		}
		return n, err
	}(l.buffer[:l.size])

	if err != nil {
		log.Errorf("Error written to stream %s", err)
		done <- true
		return nil
	}

	l.streams[stream] = true

	<-done

	delete(l.streams, stream)

	return nil
}

// Adds writer to writers to further redirect data from the read reader
func (l *Logger) Pipe(writer io.WriteCloser) error {

	var done = make(chan bool, 1)

	go func() {
		<-l.done
		done <- true
	}()

	_, err := func(p []byte) (n int, err error) {
		n, err = writer.Write(p)
		return n, err
	}(l.buffer[:l.size])

	if err != nil {
		log.Errorf("Error written to stream %s", err)
		return nil
	}

	l.writers[writer] = true

	<-done

	delete(l.writers, writer)

	return nil
}

func (l *Logger) GetBuffer() []byte {
	return l.buffer
}

func (l *Logger) Done() {
	<-l.done
}
