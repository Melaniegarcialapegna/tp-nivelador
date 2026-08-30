package protocol

//Es la interfaz publica qu usa la capa client.go

import (
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
)

// serializa apuesta, le agrega header y manda por socket
func SendBet(socket io.Writer, bet model.Bet) error { return nil }

// avisa que no se van a mandar mas apuestas
func SendEnd(socket io.Writer) error { return nil }

// espera de manera bloqueante a q llegue el mensaje del ganador y lo deserializa devolviendole al cliente una Bet
func RecvWinner(socket io.Reader) (model.Bet, error) { return model.Bet{}, nil }

// if err := safe_socket.SendAll(client.conn, []byte(lineaClientMessage)); err != nil {
// 	logger.Error("send-message", logger.Fail, messageArgs...)
// 	return err
// }

// responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)
// if err != nil {
// 	logger.Error("recv-response", logger.Fail, messageArgs...)
// 	return err
// }
