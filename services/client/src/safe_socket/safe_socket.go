package safe_socket

import (
	"errors"
	"io"
)

func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0
	for totalSent < len(bytes) { //loop continue until all bytes are sent
		n, err := socket.Write(bytes[totalSent:])
		if err != nil {
			return err
		}
		if n == 0 {
			return errors.New("connection closed after sending all data")
		}
		totalSent += n
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	totalRead := 0
	for totalRead < size { //loop continue until all bytes are read
		n, err := socket.Read(buff[totalRead:])
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, errors.New("connection closed after reading all data")
		}
		totalRead += n
	}
	return buff, nil
}
