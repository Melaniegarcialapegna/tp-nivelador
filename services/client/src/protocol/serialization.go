package protocol

import (
	"encoding/binary"
	"errors"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
)

const BIRTHDATE_LENGTH = 10 //YYYY-MM-DD
const AMOUNT_BYTES_CONST = AMOUNT_BYTES_UINT32 + AMOUNT_BYTES_UINT32 + AMOUNT_BYTES_UINT32 + BIRTHDATE_LENGTH

// Serializes a bet to a byte array
// For the fields that are dinamic in size, will be used a separator to know how many bytes to read for each field
// like long_dinamic_field_i|dinamic_field_i|
func serializeBet(bet model.Bet) []byte {
	bet_bytes := make([]byte, 0)

	//pongo a los campos fijos que van directo
	//TODO: modularizar
	agencyIdBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(agencyIdBytes, uint32(bet.AgencyId))
	bet_bytes = append(bet_bytes, agencyIdBytes...)

	documentBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(documentBytes, uint32(bet.Document))
	bet_bytes = append(bet_bytes, documentBytes...)

	numberBytes := make([]byte, AMOUNT_BYTES_UINT32)
	binary.BigEndian.PutUint32(numberBytes, uint32(bet.Number))
	bet_bytes = append(bet_bytes, numberBytes...)

	birthdateBytes := []byte(bet.Birthdate)
	bet_bytes = append(bet_bytes, birthdateBytes...)

	//pongo a los campos dinamicos que van con header longitud
	bet_bytes = append(bet_bytes, writeDynamicField(bet.FirstName)...)
	bet_bytes = append(bet_bytes, writeDynamicField(bet.LastName)...)

	return bet_bytes
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
func deserializeBet(bet_bytes []byte) (model.Bet, error) {

	//se sabe que como minimo necesitamos tener el tamaño de los campos fijos para poder deserializar
	if len(bet_bytes) < AMOUNT_BYTES_CONST {
		return model.Bet{}, errors.New("data too short to deserialize a bet") //TODO: hacer tipos errores
	}

	position := 0

	//TODO: modularizar
	agencyId := int32(binary.BigEndian.Uint32(bet_bytes[position : position+AMOUNT_BYTES_UINT32]))
	position += AMOUNT_BYTES_UINT32

	document := int32(binary.BigEndian.Uint32(bet_bytes[position : position+AMOUNT_BYTES_UINT32]))
	position += AMOUNT_BYTES_UINT32

	number := int32(binary.BigEndian.Uint32(bet_bytes[position : position+AMOUNT_BYTES_UINT32]))
	position += AMOUNT_BYTES_UINT32

	birthdate := string(bet_bytes[position : position+BIRTHDATE_LENGTH])
	position += BIRTHDATE_LENGTH

	firstName, newPosition, err := readDynamicField(bet_bytes, position)
	if err != nil {
		return model.Bet{}, err
	}
	position = newPosition

	lastName, _, err := readDynamicField(bet_bytes, position)
	if err != nil {
		return model.Bet{}, err
	}

	return model.Bet{
		AgencyId:  agencyId,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}, nil

}

func readDynamicField(representationBetBytes []byte, position int) (string, int, error) {
	if position+AMOUNT_BYTES_UINT16 > len(representationBetBytes) {
		return "", 0, errors.New("data is too short to contain field length")
	}

	//lee el largo de el str dinamico
	length := int(binary.BigEndian.Uint16(representationBetBytes[position : position+AMOUNT_BYTES_UINT16]))
	position += AMOUNT_BYTES_UINT16

	//chequea si el largo es valido
	if position+length > len(representationBetBytes) {
		return "", 0, errors.New("data is too short to contain the dynamic field")
	}

	//le el campo dinamico
	field := string(representationBetBytes[position : position+length])
	position += length

	return field, position, nil
}
