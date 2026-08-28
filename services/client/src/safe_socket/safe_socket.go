package safe_socket

import (
	"errors"
	"io"
)

func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0
	for totalSent < len(bytes) { //loop continue until all bytes are sent
		sizeSent, err := socket.Write(bytes[totalSent:])
		if err != nil {
			return err
		}
		if sizeSent == 0 {
			return errors.New("connection closed before sending all data")
		}
		totalSent += sizeSent
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	totalRead := 0
	for totalRead < size { //loop continue until all bytes are read
		sizeRead, err := socket.Read(buff[totalRead:])
		if err != nil {
			return nil, err
		}
		if sizeRead == 0 {
			return nil, errors.New("connection closed before reading all data")
		}
		totalRead += sizeRead
	}
	return buff, nil
}
