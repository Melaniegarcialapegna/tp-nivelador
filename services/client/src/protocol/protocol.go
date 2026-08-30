package protocol

//Es la interfaz publica qu usa la capa client.go

import (
	"encoding/binary"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

// TODO: ver si poner ctes en otro lado
const AMOUNT_BYTES_UINT32 = 4
const AMOUNT_BYTES_UINT16 = 2

// serializa apuesta, le agrega header y manda por socket
func SendBet(socket io.Writer, bet model.Bet) error {
	payloadBet := serializeBet(bet) //internamente maneja los campos dinamicos
	header := createHeader(len(payloadBet))

	betMessage := append(header, payloadBet...)

	//paso a siguiente capa
	if err := safe_socket.SendAll(socket, betMessage); err != nil {
		return err
	}

	return nil
}

func createHeader(payloadSize int) []byte {
	header := make([]byte, 0)
	//TODO: poner en cte
	header = append(header, '0') //tipo mensaje = enviar bet

	payloadSizeBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(payloadSizeBytes, uint32(payloadSize))
	header = append(header, payloadSizeBytes...)

	return header
}

// avisa que no se van a mandar mas apuestas
func SendEnd(socket io.Writer) error { return nil }

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
