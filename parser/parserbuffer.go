package parser

import (
	"bufio"
	"bytes"
	"io"

	"github.com/FireSpotter/gossip/log"
)

// parserBuffer is a specialized buffer for use in the parser package.
// It is written to via the non-blocking Write.
// It exposes various blocking read methods, which wait until the requested
// data is avaiable, and then return it.
type parserBuffer struct {
	io.Writer
	buffer bytes.Buffer

	// Wraps parserBuffer.pipeReader
	reader *bufio.Reader

	// Don't access this directly except when closing.
	pipeReader *io.PipeReader
}

// Create a new parserBuffer object (see struct comment for object details).
// Note that resources owned by the parserBuffer may not be able to be GCed
// until the Dispose() method is called.
func newParserBuffer() *parserBuffer {
	var pb parserBuffer
	pb.pipeReader, pb.Writer = io.Pipe()
	pb.reader = bufio.NewReader(pb.pipeReader)
	return &pb
}

// Block until the buffer contains at least one CRLF-terminated line.
// Return the line, excluding the terminal CRLF, and delete it from the buffer.
// Returns an error if the parserbuffer has been stopped.
func (pb *parserBuffer) NextLine() (response string, err error) {
	var buffer bytes.Buffer
	var data string
	var b byte

	var byteLine []byte
	for b != '\r' && b != '\n' {
		b, err = pb.reader.ReadByte()
		if err != nil {
			return
		}
		byteLine = append(byteLine, b)
	}
	if b == '\r' && pb.reader.Buffered() > 0 {
		b, err = pb.reader.ReadByte()
		if err != nil {
			return
		}
	}
	data = string(byteLine)
	buffer.WriteString(data)

	response = buffer.String()
	response = response[:len(response)-1]
	return
}

// Block until the buffer contains at least n characters.
// Return precisely those n characters, then delete them from the buffer.
func (pb *parserBuffer) NextChunk(n int) (response string, err error) {
	var data []byte
	var b byte

	for total := 0; total < n && pb.reader.Buffered() > 0; {
		b, err = pb.reader.ReadByte()
		if err != nil {
			return
		}
		data = append(data, b)
		total += 1
	}

	response = string(data)
	log.Debug("Parser buffer returns chunk '%s'", response)
	return
}

// Stop the parser buffer.
func (pb *parserBuffer) Stop() {
	pb.pipeReader.Close()
}
