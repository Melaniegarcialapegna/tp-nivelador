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
const HEADER_AMOUNT = 5

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

	//MANDO 4 bytes de longitud en 0 para respetar header
	payloadSizeBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(payloadSizeBytes, uint32(0))
	messageEnd = append(messageEnd, payloadSizeBytes...)

	if err := safe_socket.SendAll(socket, messageEnd); err != nil {
		logger.Error("send-message", logger.Fail) //TODO: ver que mas agregarle al log y si es correcto ponerlo aca o con la capa de arriba basta xd
		return err
	}
	return nil
}

// espera de manera bloqueante a q llegue el mensaje del ganador y lo deserializa devolviendole al cliente una Bet
// aca manejo el tema de hasta cuando me llegan, la app se abstrae de como se yo que termine de recibir ganadores
func ReceiveWinners(socket io.Reader) ([]model.Bet, error) {

	//idea:
	//leer header entero
	//si 0 : leo resto y si 1 devuelvo

	winnersBets := make([]model.Bet, 0)

	//TODO: modularizar
	headerBuffer, err := safe_socket.RecvAll(socket, HEADER_AMOUNT)
	if err != nil {
		logger.Error("recv-response", logger.Fail) //TODO
		return []model.Bet{}, err
	}

	for headerBuffer[0] == byte('0') {
		lenghtPayload := binary.BigEndian.Uint32(headerBuffer[AMOUNT_BYTES_UINT8:HEADER_AMOUNT])

		winnerBetBytes, err := safe_socket.RecvAll(socket, int(lenghtPayload))
		if err != nil {
			logger.Error("recv-response", logger.Fail) //TODO
			return []model.Bet{}, err
		}

		winnerBet, err := deserializeBet(winnerBetBytes)
		if err != nil {
			logger.Error("deserialize-bet", logger.Fail) //TODO
			return []model.Bet{}, err
		}

		winnersBets = append(winnersBets, winnerBet)

		headerBuffer, err = safe_socket.RecvAll(socket, HEADER_AMOUNT)
		if err != nil {
			logger.Error("recv-response", logger.Fail) //TODO
			return []model.Bet{}, err
		}
	}

	return winnersBets, nil
}
