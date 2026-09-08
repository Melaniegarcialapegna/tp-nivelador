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

type Protocol struct {
	conn io.ReadWriter
}

func NewProtocol(conn io.ReadWriter) *Protocol {
	return &Protocol{conn: conn}
}

// Sends a batch of bets to the server
func (p *Protocol) SendBetBatch(bets []model.Bet) error {
	batchMessage := make([]byte, EMPTY_SLICE)

	batchMessageWithoutHeader := p.makeBatchMessageWithoutHeader(bets)

	header := p.createHeader(len(batchMessageWithoutHeader))

	batchMessage = append(batchMessage, header...)
	batchMessage = append(batchMessage, batchMessageWithoutHeader...)

	//to the next layer
	if err := safe_socket.SendAll(p.conn, batchMessage); err != nil {
		logger.Error("send-message", logger.Fail)
		return err
	}
	return nil
}

// Waits for an ACK message from the server and returns true if the ACK is OK, false if the ACK is FAIL, and an error if there was an error receiving the message or if the message type is unexpected.
func (p *Protocol) ReceiveAck() (bool, error) {
	action := "receive-ack"
	message, err := safe_socket.RecvAll(p.conn, ACK_MESSAGE_SIZE_BYTES)

	if err != nil {
		logger.Error(action, logger.Fail)
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
func (p *Protocol) SendEnd() error {
	action := "send-end"
	messageEnd := make([]byte, EMPTY_SLICE)

	messageEnd = append(messageEnd, p.getBytesOfTypeMessage(MESSAGE_TYPE_END)...)

	//MANDO 4 bytes de longitud en 0 para respetar header
	messageEnd = append(messageEnd, p.getFieldBytesForUint32(uint32(EMPTY_MESSAGE))...)

	if err := safe_socket.SendAll(p.conn, messageEnd); err != nil {
		logger.Error(action, logger.Fail)
		return err
	}
	return nil
}

// Waits for the server to send the winners and returns a slice of bets representing the winners. It will keep receiving bets until it receives a message of type END.
func (p *Protocol) ReceiveWinners() ([]model.Bet, error) {
	const action = "recv-winners"
	winnersBets := make([]model.Bet, EMPTY_SLICE)

	headerBuffer, err := p.receiveHeader(action)
	if err != nil {
		logger.Error(action, logger.Fail)
		return []model.Bet{}, err
	}

	for headerBuffer[0] == byte(MESSAGE_TYPE_BET) {

		lenghtBet := binary.BigEndian.Uint32(headerBuffer[TYPE_MESSAGE_SIZE_BYTES:HEADER_SIZE_BYTES])

		winnerBet, err := p.receiveBet(lenghtBet)
		if err != nil {
			logger.Error(action, logger.Fail)
			return []model.Bet{}, err
		}

		winnersBets = append(winnersBets, winnerBet)

		headerBuffer, err = p.receiveHeader(action)
		if err != nil {
			logger.Error(action, logger.Fail)
			return []model.Bet{}, err
		}

	}

	if p.checkEndOfBets(action, headerBuffer) == false {
		return []model.Bet{}, errors.New("unexpected message type received")
	}

	return winnersBets, nil
}

func (p *Protocol) receiveHeader(action string) ([]byte, error) {
	headerBuffer, err := safe_socket.RecvAll(p.conn, HEADER_SIZE_BYTES)
	if err != nil {
		logger.Error(action, logger.Fail)
		return nil, err
	}
	return headerBuffer, nil
}

func (p *Protocol) receiveBet(lenghtBet uint32) (model.Bet, error) {
	action := "recv-bet"
	winnerBetBytes, err := safe_socket.RecvAll(p.conn, int(lenghtBet))
	if err != nil {
		logger.Error(action, logger.Fail)
		return model.Bet{}, err
	}

	winnerBet, err := p.deserializeBet(winnerBetBytes)
	if err != nil {
		logger.Error(action, logger.Fail)
		return model.Bet{}, err
	}
	return winnerBet, nil
}

func (p *Protocol) makeBatchMessageWithoutHeader(bets []model.Bet) []byte {
	batchMessage := make([]byte, EMPTY_SLICE)
	for _, bet := range bets {
		betBytes := p.serializeBet(bet) //intern handle dynamic fields

		lenhgtBetBytes := make([]byte, LENGTH_FIELD_SIZE_BYTES)
		binary.BigEndian.PutUint32(lenhgtBetBytes, uint32(len(betBytes)))

		batchMessage = append(batchMessage, lenhgtBetBytes...)
		batchMessage = append(batchMessage, betBytes...)
	}
	return batchMessage
}

func (p *Protocol) createHeader(batchSize int) []byte {
	header := make([]byte, EMPTY_SLICE)

	//put the type of message in the header
	header = append(header, p.getBytesOfTypeMessage(MESSAGE_TYPE_BET)...)

	//put the length of batch in the header
	header = append(header, p.getFieldBytesForUint32(uint32(batchSize))...)

	return header
}

func (p *Protocol) getBytesOfTypeMessage(messageType int) []byte {
	typeMessageBytes := make([]byte, TYPE_MESSAGE_SIZE_BYTES)
	typeMessageBytes[0] = byte(messageType)
	return typeMessageBytes
}

func (p *Protocol) checkEndOfBets(action string, headerBuffer []byte) bool {
	if headerBuffer[0] != byte(MESSAGE_TYPE_END) {
		logger.Error(action, logger.Fail, "unexpected-message-type")
		return false
	}
	return true
}
