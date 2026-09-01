package protocol

//Es la interfaz publica qu usa la capa client.go

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

// Sends a batch of bets to the server
func SendBetBatch(socket io.Writer, bets []model.Bet) error {
	batchMessage := make([]byte, EMPTY_SLICE)

	batchMessageWithoutHeader := makeBatchMessageWithoutHeader(bets)

	header := createHeader(len(batchMessageWithoutHeader))

	batchMessage = append(batchMessage, header...)
	batchMessage = append(batchMessage, batchMessageWithoutHeader...)

	//to the next layer
	if err := safe_socket.SendAll(socket, batchMessage); err != nil {
		logger.Error("send-message", logger.Fail)
		return err
	}
	return nil
}

func makeBatchMessageWithoutHeader(bets []model.Bet) []byte {
	batchMessage := make([]byte, EMPTY_SLICE)
	for _, bet := range bets {
		betBytes := serializeBet(bet) //intern handle dynamic fields

		lenhgtBetBytes := make([]byte, LENGTH_FIELD_SIZE_BYTES)
		binary.BigEndian.PutUint32(lenhgtBetBytes, uint32(len(betBytes)))

		batchMessage = append(batchMessage, lenhgtBetBytes...)
		batchMessage = append(batchMessage, betBytes...)
	}
	return batchMessage
}

func createHeader(batchSize int) []byte {
	header := make([]byte, EMPTY_SLICE)

	//put the type of message in the header
	header = append(header, getBytesOfTypeMessage(MESSAGE_TYPE_BET)...)

	//put the length of batch in the header
	header = append(header, getFieldBytesForUint32(uint32(batchSize))...)

	return header
}

func getBytesOfTypeMessage(messageType int) []byte {
	typeMessageBytes := make([]byte, TYPE_MESSAGE_SIZE_BYTES)
	typeMessageBytes[0] = byte(messageType)
	return typeMessageBytes
}

// ---

// Waits for an ACK message from the server and returns true if the ACK is OK, false if the ACK is FAIL, and an error if there was an error receiving the message or if the message type is unexpected.
func ReceiveAck(socket io.Reader) (bool, error) {

	message, err := safe_socket.RecvAll(socket, ACK_MESSAGE_SIZE_BYTES)

	if err != nil {
		logger.Error("recv-ack", logger.Fail)
		return false, err
	}

	if message[0] == byte(MESSAGE_TYPE_ACK_OK) {
		return true, nil
	}
	if message[0] == byte(MESSAGE_TYPE_ACK_FAIL) {
		return false, nil
	}

	return false, errors.New("unexpected message type received")
}

// Sends an END message to the server to indicate that no more bets will be sent
func SendEnd(socket io.Writer) error {
	messageEnd := make([]byte, EMPTY_SLICE)

	messageEnd = append(messageEnd, getBytesOfTypeMessage(MESSAGE_TYPE_END)...)

	//MANDO 4 bytes de longitud en 0 para respetar header
	messageEnd = append(messageEnd, getFieldBytesForUint32(uint32(EMPTY_MESSAGE))...)

	if err := safe_socket.SendAll(socket, messageEnd); err != nil {
		logger.Error("send-message", logger.Fail)
		return err
	}
	return nil
}

// Waits for the server to send the winners and returns a slice of bets representing the winners. It will keep receiving bets until it receives a message of type END.
func ReceiveWinners(socket io.Reader) ([]model.Bet, error) {
	winnersBets := make([]model.Bet, EMPTY_SLICE)

	headerBuffer, err := receiveHeader(socket)
	if err != nil {
		logger.Error(ACTION_RECEIVE_WINNERS, logger.Fail)
		return []model.Bet{}, err
	}

	for headerBuffer[0] == byte(MESSAGE_TYPE_BET) {

		lenghtBet := binary.BigEndian.Uint32(headerBuffer[TYPE_MESSAGE_SIZE_BYTES:HEADER_SIZE_BYTES])

		winnerBet, err := receiveBet(socket, lenghtBet)
		if err != nil {
			logger.Error(ACTION_RECEIVE_WINNERS, logger.Fail)
			return []model.Bet{}, err
		}

		winnersBets = append(winnersBets, winnerBet)

		headerBuffer, err = receiveHeader(socket)
		if err != nil {
			logger.Error(ACTION_RECEIVE_WINNERS, logger.Fail)
			return []model.Bet{}, err
		}

	}

	if checkEndOfBets(headerBuffer) == false {
		return []model.Bet{}, errors.New("unexpected message type received")
	}

	return winnersBets, nil
}

func receiveHeader(socket io.Reader) ([]byte, error) {
	headerBuffer, err := safe_socket.RecvAll(socket, HEADER_SIZE_BYTES)
	if err != nil {
		logger.Error("recv-header", logger.Fail)
		return nil, err
	}
	return headerBuffer, nil
}

func receiveBet(socket io.Reader, lenghtBet uint32) (model.Bet, error) {
	winnerBetBytes, err := safe_socket.RecvAll(socket, int(lenghtBet))
	if err != nil {
		logger.Error(ACTION_RECEIVE_WINNERS, logger.Fail)
		return model.Bet{}, err
	}

	winnerBet, err := deserializeBet(winnerBetBytes)
	if err != nil {
		logger.Error("deserialize-bet", logger.Fail)
		return model.Bet{}, err
	}
	return winnerBet, nil
}

func checkEndOfBets(headerBuffer []byte) bool {
	if headerBuffer[0] != byte(MESSAGE_TYPE_END) {
		logger.Error("recv-response", logger.Fail, "unexpected-message-type")
		return false
	}
	return true
}
