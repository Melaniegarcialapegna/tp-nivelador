package protocol

import (
	"encoding/binary"
	"errors"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
)

// Serializes a bet to a byte array
// For the fields that are dinamic in size, will be used a separator to know how many bytes to read for each field
// like long_dinamic_field_i|dinamic_field_i|
func serializeBet(bet model.Bet) []byte {
	betBytes := make([]byte, 0)

	//put fixed fields
	betBytes = append(betBytes, getFieldBytesForUint32(uint32(bet.AgencyId))...)
	betBytes = append(betBytes, getFieldBytesForUint32(uint32(bet.Document))...)
	betBytes = append(betBytes, getFieldBytesForUint32(uint32(bet.Number))...)

	birthdateBytes := []byte(bet.Birthdate)
	betBytes = append(betBytes, birthdateBytes...)

	//put dinamic fields that have a length prefix
	betBytes = append(betBytes, getDynamicField(bet.FirstName)...)
	betBytes = append(betBytes, getDynamicField(bet.LastName)...)

	return betBytes
}

func getFieldBytesForUint32(field uint32) []byte {
	fieldBytes := make([]byte, FIXED_FIELDS_SIZE_BYTES)
	binary.BigEndian.PutUint32(fieldBytes, field)
	return fieldBytes
}

func getDynamicField(field string) []byte {
	//field in bytes
	fieldBytes := []byte(field)

	//length of the field in bytes
	lengthFieldBytes := make([]byte, DYNAMIC_FIELD_LENGTH_SIZE_BYTES)
	binary.BigEndian.PutUint16(lengthFieldBytes, uint16(len(fieldBytes)))

	return append(lengthFieldBytes, fieldBytes...)

}

// Deserializes a byte array to a bet
func deserializeBet(betBytes []byte) (model.Bet, error) {
	if len(betBytes) < AMOUNT_BYTES_CONST {
		return model.Bet{}, errors.New("data too short to deserialize a bet") //TODO: hacer tipos errores
	}

	position := 0

	agencyId, position := getInt32FromBytes(betBytes, position)
	document, position := getInt32FromBytes(betBytes, position)
	number, position := getInt32FromBytes(betBytes, position)

	birthdate := string(betBytes[position : position+BIRTHDATE_LENGTH])
	position += BIRTHDATE_LENGTH

	firstName, position, err := readDynamicField(betBytes, position)
	if err != nil {
		return model.Bet{}, err
	}

	lastName, _, err := readDynamicField(betBytes, position)
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

func getInt32FromBytes(bytes []byte, position int) (int32, int) {
	return int32(binary.BigEndian.Uint32(bytes[position : position+FIXED_FIELDS_SIZE_BYTES])), position + FIXED_FIELDS_SIZE_BYTES
}

func readDynamicField(betBytes []byte, position int) (string, int, error) {
	if position+DYNAMIC_FIELD_LENGTH_SIZE_BYTES > len(betBytes) {
		return "", 0, errors.New("data is too short to contain field length")
	}

	//read the length of the dynamic field
	length := int(binary.BigEndian.Uint16(betBytes[position : position+DYNAMIC_FIELD_LENGTH_SIZE_BYTES]))
	position += DYNAMIC_FIELD_LENGTH_SIZE_BYTES

	if position+length > len(betBytes) {
		return "", 0, errors.New("data is too short to contain the dynamic field")
	}

	//read the dynamic field based on the length
	field := string(betBytes[position : position+length])
	position += length

	return field, position, nil
}
