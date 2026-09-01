package protocol

// Constants for serialization and protocol

// Protocol
const (
	MESSAGE_TYPE_BET      = 0 //context of bets
	MESSAGE_TYPE_END      = 1 //finish of sending bets
	MESSAGE_TYPE_ACK_OK   = 2 //ack of reception without error
	MESSAGE_TYPE_ACK_FAIL = 3 //ack of reception with error
)
const TYPE_MESSAGE_SIZE_BYTES = 1
const ACK_MESSAGE_SIZE_BYTES = 1

const LENGTH_FIELD_SIZE_BYTES = 4

const FIXED_FIELDS_SIZE_BYTES = 4
const BIRTHDATE_LENGTH = 10 //YYYY-MM-DD
const DYNAMIC_FIELD_LENGTH_SIZE_BYTES = 2

const AMOUNT_BYTES_CONST = FIXED_FIELDS_SIZE_BYTES*3 + BIRTHDATE_LENGTH // agency_id + document + number + birthdate
const HEADER_SIZE_BYTES = TYPE_MESSAGE_SIZE_BYTES + LENGTH_FIELD_SIZE_BYTES

const EMPTY_MESSAGE = 0
