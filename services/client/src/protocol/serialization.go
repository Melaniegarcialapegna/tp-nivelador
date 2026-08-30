package protocol

import "github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"

// serializa la apuesta (otro archivo)
func serializeBet(bet model.Bet) []byte {
	return []byte{}
}

func deserializeBet(data []byte) model.Bet { return model.Bet{} }
