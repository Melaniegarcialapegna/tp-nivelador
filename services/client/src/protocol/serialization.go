package protocol

import (
	"encoding/binary"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
)

// Serializes a bet to a byte array
// For the fields that are dinamic in size, will be used a separator to know how many bytes to read for each field
// like long_dinamic_field_i|dinamic_field_i|
func serializeBet(bet model.Bet) []byte {
	representationBetBytes := make([]byte, 0)

	//pongo a los campos fijos que van directo
	//TODO: modularizar
	agencyIdBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(agencyIdBytes, uint32(bet.AgencyId))
	representationBetBytes = append(representationBetBytes, agencyIdBytes...)

	documentBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(documentBytes, uint32(bet.Document))
	representationBetBytes = append(representationBetBytes, documentBytes...)

	numberBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(numberBytes, uint32(bet.Number))
	representationBetBytes = append(representationBetBytes, numberBytes...)

	birthdateBytes := []byte(bet.Birthdate)
	representationBetBytes = append(representationBetBytes, birthdateBytes...)

	//pongo a los campos dinamicos que van con header longitud
	representationBetBytes = append(representationBetBytes, writeDynamicField(bet.FirstName)...)
	representationBetBytes = append(representationBetBytes, writeDynamicField(bet.LastName)...)

	return representationBetBytes
}

func writeDynamicField(field string) []byte {
	//campo en bytes
	fieldBytes := []byte(field)

	//longitud campo en bytes
	lengthFieldBytes := make([]byte, AMOUNT_BYTES_UINT16)
	binary.BigEndian.PutUint16(lengthFieldBytes, uint16(len(fieldBytes)))

	return append(lengthFieldBytes, fieldBytes...)

}

// Deserializes a byte array to a bet
func deserializeBet(data []byte) model.Bet { return model.Bet{} }

func readDynamicField(data []byte) {}
