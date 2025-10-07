package mcp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type messageReader struct {
	reader *bufio.Reader
}

func newMessageReader(r io.Reader) *messageReader {
	return &messageReader{reader: bufio.NewReader(r)}
}

func (mr *messageReader) Next() ([]byte, error) {
	var (
		contentLength int
		sawHeader     bool
	)
	for {
		line, err := mr.reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && line == "" {
				return nil, io.EOF
			}
			if err == io.EOF {
				return nil, fmt.Errorf("unexpected EOF while reading headers")
			}
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if !sawHeader {
				continue
			}
			break
		}

		sawHeader = true
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header line: %q", line)
		}
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if strings.EqualFold(name, "Content-Length") {
			length, err := strconv.Atoi(value)
			if err != nil || length < 0 {
				return nil, fmt.Errorf("invalid Content-Length value: %q", value)
			}
			contentLength = length
		}
	}

	if !sawHeader {
		return nil, io.EOF
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(mr.reader, payload); err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	return payload, nil
}

type messageWriter struct {
	writer *bufio.Writer
}

func newMessageWriter(w io.Writer) *messageWriter {
	return &messageWriter{writer: bufio.NewWriter(w)}
}

func (mw *messageWriter) Write(payload []byte) error {
	if _, err := fmt.Fprintf(mw.writer, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return err
	}
	if _, err := mw.writer.Write(payload); err != nil {
		return err
	}
	return mw.writer.Flush()
}
