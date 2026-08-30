package protocol

//Es la interfaz publica qu usa la capa client.go

import (
	"encoding/binary"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

// TODO: ver si poner ctes en otro lado
const AMOUNT_BYTES_UINT8 = 1
const AMOUNT_BYTES_UINT16 = 2
const AMOUNT_BYTES_UINT32 = 4

// serializa apuesta, le agrega header y manda por socket
func SendBet(socket io.Writer, bet model.Bet) error {
	payloadBet := serializeBet(bet) //internamente maneja los campos dinamicos
	header := createHeader(len(payloadBet))

	betMessage := append(header, payloadBet...)

	//paso a siguiente capa
	if err := safe_socket.SendAll(socket, betMessage); err != nil {
		logger.Error("send-message", logger.Fail) //TODO: ver que mas agregarle al log y si es correcto ponerlo aca o con la capa de arriba basta xd
		return err
	}
	return nil
}

func createHeader(payloadSize int) []byte {
	header := make([]byte, 0) //TODO: poner 0 en cte

	typeMessageBytes := make([]byte, AMOUNT_BYTES_UINT8)
	typeMessageBytes[0] = byte('0')
	header = append(header, typeMessageBytes...)

	payloadSizeBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(payloadSizeBytes, uint32(payloadSize))
	header = append(header, payloadSizeBytes...)

	return header
}

// avisa que no se van a mandar mas apuestas
func SendEnd(socket io.Writer) error {
	messageEnd := make([]byte, 0) //TODO: poner 0 en cte

	typeMessageBytes := make([]byte, AMOUNT_BYTES_UINT8)
	typeMessageBytes[0] = byte('1')
	messageEnd = append(messageEnd, typeMessageBytes...)

	if err := safe_socket.SendAll(socket, messageEnd); err != nil {
		logger.Error("send-message", logger.Fail) //TODO: ver que mas agregarle al log y si es correcto ponerlo aca o con la capa de arriba basta xd
		return err
	}
	return nil
}

// espera de manera bloqueante a q llegue el mensaje del ganador y lo deserializa devolviendole al cliente una Bet
func ReceiveWinners(socket io.Reader) ([]model.Bet, error) { return []model.Bet{}, nil }

// if err := safe_socket.SendAll(client.conn, []byte(lineaClientMessage)); err != nil {
// 	logger.Error("send-message", logger.Fail, messageArgs...)
// 	return err
// }

// responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)
// if err != nil {
// 	logger.Error("recv-response", logger.Fail, messageArgs...)
// 	return err
// }
