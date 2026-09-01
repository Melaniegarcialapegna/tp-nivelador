package client

import (
	"bufio"
	"fmt"
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
const AMOUNT_OF_FIELDS_PER_BET = 5
const EMPTY_SLICE = 0

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

	batch := make([]model.Bet, EMPTY_SLICE, client.config.BatchSize)

	for scanner.Scan() {
		lineBet := scanner.Text()

		bet, err := parseBetFromCSVLine(lineBet, client.config.AgencyId)
		if err != nil {
			return err
		}

		batch = append(batch, bet)

		if len(batch) == client.config.BatchSize {
			if err := sendBatch(client, batch); err != nil {
				return err
			}
			batch = batch[:EMPTY_SLICE] //drop
		}
	}

	//Case when the scanner encounters an error while reading the input file
	if err := scanner.Err(); err != nil {
		logger.Error("read-input-file", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}

	//Case when the last batch is not full, but there are still bets to send
	if len(batch) > EMPTY_SLICE {
		if err := sendBatch(client, batch); err != nil {
			return err
		}
	}

	//Send to the protocol that the client has finished sending bets
	if err := protocol.SendEnd(client.conn); err != nil {
		logger.Error("send-end", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}

	//Waits for the winners bets from the server
	winnersBets, err := protocol.ReceiveWinners(client.conn)
	if err != nil {
		logger.Error("receive-winners", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}

	if err := persistWinnersToFile(winnersBets, outputFile, client.config.AgencyId); err != nil {
		return err
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}

func persistWinnersToFile(winnersBets []model.Bet, outputFile *os.File, agencyId string) error {
	for _, winnerBet := range winnersBets {
		lineBet := winnerBet.FirstName + "," + winnerBet.LastName + "," + strconv.Itoa(int(winnerBet.Document)) + "," + winnerBet.Birthdate + "," + strconv.Itoa(int(winnerBet.Number)) + "\n"
		if _, err := outputFile.WriteString(lineBet); err != nil {
			logger.Error("write-output", logger.Fail, "agency-id", agencyId, "error", err)
			return err
		}
	}
	return nil
}

func sendBatch(client *Client, batch []model.Bet) error {
	if err := protocol.SendBetBatch(client.conn, batch); err != nil {
		logger.Error("send-bet", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}

	//waits for the ack from the server
	success, err := protocol.ReceiveAck(client.conn)

	if !success {
		logger.Error("receive-ack", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}
	if err != nil {
		logger.Error("receive-ack", logger.Fail, "agency-id", client.config.AgencyId, "error", err)
		return err
	}
	return nil
}

func parseBetFromCSVLine(line string, agencyIdStr string) (model.Bet, error) {
	fields := strings.Split(line, ",")
	if len(fields) != AMOUNT_OF_FIELDS_PER_BET {
		return model.Bet{}, fmt.Errorf("expected %d fields, got %d", AMOUNT_OF_FIELDS_PER_BET, len(fields))
	}

	agencyId, err := parsefields(agencyIdStr, 0, "parse-agency-id")
	if err != nil {
		return model.Bet{}, err
	}

	document, err := parsefields(fields[2], int(agencyId), "parse-document")
	if err != nil {
		return model.Bet{}, err
	}

	number, err := parsefields(fields[4], int(agencyId), "parse-number")
	if err != nil {
		return model.Bet{}, err
	}

	bet := model.Bet{
		AgencyId:  int32(agencyId),
		FirstName: fields[0],
		LastName:  fields[1],
		Document:  int32(document),
		Birthdate: fields[3],
		Number:    int32(number),
	}

	return bet, nil

}

func parsefields(fieldStr string, agencyId int, action string) (int32, error) {
	field, err := strconv.Atoi(fieldStr)
	if err != nil {
		logger.Error(action, logger.Fail, "agency-id", agencyId, "error", err)
		return 0, err
	}
	return int32(field), nil
}
