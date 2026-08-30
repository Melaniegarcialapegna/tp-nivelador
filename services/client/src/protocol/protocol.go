package protocol

//Es la interfaz publica qu usa la capa client.go

import "io"

// serializa apuesta, le agrega header y manda por socket
func SendBet(socket io.Writer, bet Bet, isLast bool) error { return nil }

// espera de manera bloqueante a q llegue el mensaje del ganador y lo deserializa devolviendole al cliente una Bet
func RecvWinner(socket io.Reader) (Bet, error) { return Bet{}, nil }
