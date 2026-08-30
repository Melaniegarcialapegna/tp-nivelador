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

const ECHO_CLIENT_BUFFER_SIZE = 512
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
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
	const mainAction = "test-echo-server" // TODO: cambiar
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

	for scanner.Scan() {
		lineBet := scanner.Text()
		messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		logger.Info(mainAction, logger.InProgress, messageArgs...)

		//Armo Bet
		//TODO : Modularizar en una funcion
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

		//Se le pasa al protocolo para que lo envie
		protocol.SendBet(client.conn, bet)

		messageId++
	}

	//Caso en el que sale del loop no porque se termina de leer el archivo sino porque ocurrio un error al leer
	if err := scanner.Err(); err != nil {
		logger.Error("read-input-file", logger.Fail, "error", err)
		return err
	}

	//Se le avisa al protocolo que no se van a mandar mas apuestas
	protocol.SendEnd(client.conn)
	//Se espera la rta del ganador y se persiste en archivo

	//Se persiste en archivo
	// if _, err := outputFile.WriteString(string(responseBuffer) + "\n"); err != nil {
	// 	logger.Error("write-output", logger.Fail, messageArgs...)
	// 	return err
	// }

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
