
from lottery import Bet 

AMOUNT_BYTES_UINT32 = 4 #TODO : cambiar xd
AMOUNT_BYTES_UINT16 = 2 #TODO : cambiar xd
BIRTHDATE_LENGTH = 10 #YYYY-MM-DD
AMOUNT_BYTES_CONST = AMOUNT_BYTES_UINT32 + AMOUNT_BYTES_UINT32 + AMOUNT_BYTES_UINT32 + BIRTHDATE_LENGTH

# Serializes a bet to a byte array
# For the fields that are dinamic in size, will be used a separator to know how many bytes to read for each field
# like long_dinamic_field_i|dinamic_field_i|
def serialize_bet(bet: Bet) -> bytes:
    #pongo los campos fijos
    bet_bytes = bet.agency_id.to_bytes(AMOUNT_BYTES_UINT32, byteorder='big')
    bet_bytes += bet.document.to_bytes(AMOUNT_BYTES_UINT32, byteorder='big')
    bet_bytes += bet.number.to_bytes(AMOUNT_BYTES_UINT32, byteorder='big')
    bet_bytes += str(bet.birthdate).encode('utf-8')

    #pongo los campos dinmicos
    bet_bytes += get_dynamic_field(bet.first_name)
    bet_bytes += get_dynamic_field(bet.last_name)

    return bet_bytes

def get_dynamic_field(field: str) -> bytes:
    #campos en bytes
    field_bytes = field.encode('utf-8')
    #longitud campo
    field_length = len(field_bytes)
    field_length_bytes = field_length.to_bytes(AMOUNT_BYTES_UINT16, byteorder='big')

    return field_length_bytes + field_bytes

#Deserializes a byte array to a bet
def deserialize_bet(bet_bytes: bytes) -> Bet:
    if len(bet_bytes) < AMOUNT_BYTES_CONST:
        raise ValueError("Data too short to deserialize a bet")

    position = 0

    #TODO: modularizar
    agency_id = int.from_bytes(bet_bytes[position:position + AMOUNT_BYTES_UINT32], byteorder='big')
    position += AMOUNT_BYTES_UINT32

    document = int.from_bytes(bet_bytes[position:position + AMOUNT_BYTES_UINT32], byteorder='big')
    position += AMOUNT_BYTES_UINT32

    number = int.from_bytes(bet_bytes[position:position + AMOUNT_BYTES_UINT32], byteorder='big')
    position += AMOUNT_BYTES_UINT32

    birthdate = bet_bytes[position:position + BIRTHDATE_LENGTH].decode('utf-8')
    position += BIRTHDATE_LENGTH

    first_name, newPosition = read_dynamic_field(bet_bytes, position)
    position = newPosition

    last_name, newPosition = read_dynamic_field(bet_bytes, position)

    return Bet(
        agency_id=agency_id,
        first_name=first_name,
        last_name=last_name,
        document=document,
        birthdate=birthdate,
        number=number)


def read_dynamic_field(bet_bytes: bytes, position: int) -> (str, int):
    if position + AMOUNT_BYTES_UINT16 > len(bet_bytes):
        raise ValueError("Data is too short to contain field length")

    length = int.from_bytes(bet_bytes[position:position + AMOUNT_BYTES_UINT16], byteorder='big')
    position += AMOUNT_BYTES_UINT16

    if position + length > len(bet_bytes):
        raise ValueError("Data is too short to contain the dynamic field")

    field = bet_bytes[position:position + length].decode('utf-8')
    position += length

    return field, position
	
