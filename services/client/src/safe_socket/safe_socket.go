package safe_socket

import (
	"io"
)

func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0
	for totalSent < len(bytes) { //loop continue until all bytes are sent
		sizeSent, err := socket.Write(bytes[totalSent:])
		if err != nil {
			return err
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
		totalRead += sizeRead
	}
	return buff, nil
}
