package client

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server" //TODO: cambiar
	defer client.conn.Close()

	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail, "error", err)
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("open-output-file", logger.Fail, "error", err)
		return err
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)

	messageId := 0

	batch := make([]model.Bet, 0, client.config.BatchSize)

	for scanner.Scan() {
		lineBet := scanner.Text()
		messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		// logger.Info(mainAction, logger.InProgress, messageArgs...)

		//Armo Bet
		//TODO : Modularizar todo xd
		fieldsBet := strings.Split(lineBet, ",")

		agencyId, err := strconv.Atoi(client.config.AgencyId)
		if err != nil {
			logger.Error("parse-agency-id", logger.Fail, messageArgs...)
			return err
		}

		document, err := strconv.Atoi(fieldsBet[2])
		if err != nil {
			logger.Error("parse-document", logger.Fail, messageArgs...)
			return err
		}

		number, err := strconv.Atoi(fieldsBet[4])
		if err != nil {
			logger.Error("parse-number", logger.Fail, messageArgs...)
			return err
		}

		bet := model.Bet{
			AgencyId:  int32(agencyId),
			FirstName: fieldsBet[0],
			LastName:  fieldsBet[1],
			Document:  int32(document),
			Birthdate: fieldsBet[3],
			Number:    int32(number),
		}

		batch = append(batch, bet)

		if len(batch) == client.config.BatchSize {
			//Se le pasa al protocolo para que lo envie
			if err := protocol.SendBetBatch(client.conn, batch); err != nil {
				logger.Error("send-bet", logger.Fail, messageArgs...)
				return err
			}

			//espera el ack del server
			success, err := protocol.ReceiveAck(client.conn)

			if !success {
				logger.Error("receive-ack", logger.Fail, messageArgs...)
				return err
			}
			if err != nil {
				logger.Error("receive-ack", logger.Fail, messageArgs...)
				return err
			}

			batch = batch[:0]
		}

		messageId++
	}

	//Caso en el que sale del loop no porque se termina de leer el archivo sino porque ocurrio un error al leer
	if err := scanner.Err(); err != nil {
		//TODO : sacar codigo rep del "agency-id y error"
		logger.Error("read-input-file", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}

	//Si quedaron apuestas en el batch que no se enviaron porque no se lleno el batch
	if len(batch) > 0 {
		if err := protocol.SendBetBatch(client.conn, batch); err != nil {
			logger.Error("send-bet", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
			return err
		}

		//espera el ack del server
		success, err := protocol.ReceiveAck(client.conn)

		if !success {
			logger.Error("receive-ack", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
			return err
		}
		if err != nil {
			logger.Error("receive-ack", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
			return err
		}
	}

	//Se le avisa al protocolo que no se van a mandar mas apuestas
	if err := protocol.SendEnd(client.conn); err != nil {
		logger.Error("send-end", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}

	//Se espera la rta del ganador y se persiste en archivo
	//TODO : Modularizar todo
	winnersBets, err := protocol.ReceiveWinners(client.conn)
	if err != nil {
		logger.Error("receive-winners", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}

	for _, winnerBet := range winnersBets {
		lineBet := winnerBet.FirstName + "," + winnerBet.LastName + "," + strconv.Itoa(int(winnerBet.Document)) + "," + winnerBet.Birthdate + "," + strconv.Itoa(int(winnerBet.Number)) + "\n"
		if _, err := outputFile.WriteString(lineBet); err != nil {
			logger.Error("write-output", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
			return err
		}
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
