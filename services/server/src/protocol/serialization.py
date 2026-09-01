from lottery import Bet 

FIXED_FIELDS_SIZE_BYTES = 4 
BIRTHDATE_LENGTH = 10 #YYYY-MM-DD

DYNAMIC_FIELD_LENGTH_SIZE_BYTES = 2 

AMOUNT_BYTES_CONST = FIXED_FIELDS_SIZE_BYTES * 3 + BIRTHDATE_LENGTH  # agency_id + document + number + birthdate

# Serializes a bet to a byte array
# For the fields that are dinamic in size, will be used a separator to know how many bytes to read for each field 
# long_dinamic_field_i | dinamic_field_i |
def serialize_bet(bet: Bet) -> bytes:
    #put fixed fields 
    bet_bytes = _get_field_bytes(bet.agency_id)
    bet_bytes += _get_field_bytes(bet.document)
    bet_bytes += _get_field_bytes(bet.number)
    bet_bytes += str(bet.birthdate).encode('utf-8')

    #put dinamic fields
    bet_bytes += _get_dynamic_field(bet.first_name)
    bet_bytes += _get_dynamic_field(bet.last_name)

    return bet_bytes

def _get_field_bytes(field) -> bytes:
    return field.to_bytes(FIXED_FIELDS_SIZE_BYTES, byteorder='big')

def _get_dynamic_field(field: str) -> bytes:
    #convert field to bytes
    field_bytes = field.encode('utf-8')

    #length of the field in bytes
    field_length = len(field_bytes)
    field_length_bytes = field_length.to_bytes(DYNAMIC_FIELD_LENGTH_SIZE_BYTES, byteorder='big')

    return field_length_bytes + field_bytes #header of the bet

#--

#Deserializes a byte array to a bet
def deserialize_bet(bet_bytes: bytes) -> Bet:
    if len(bet_bytes) < AMOUNT_BYTES_CONST:
        raise ValueError("Data too short to deserialize a bet")

    position = 0

    agency_id , position = _get_int_from_field(bet_bytes, position, FIXED_FIELDS_SIZE_BYTES)
    document , position = _get_int_from_field(bet_bytes, position, FIXED_FIELDS_SIZE_BYTES)
    number , position = _get_int_from_field(bet_bytes, position, FIXED_FIELDS_SIZE_BYTES)

    birthdate , position = _get_str_from_field(bet_bytes, position, BIRTHDATE_LENGTH)

    first_name, position = _read_dynamic_field(bet_bytes, position)

    last_name, _ = _read_dynamic_field(bet_bytes, position)

    return Bet(
        agency_id=agency_id,
        first_name=first_name,
        last_name=last_name,
        document=document,
        birthdate=birthdate,
        number=number)

def _read_dynamic_field(bet_bytes: bytes, position: int):
    if position + DYNAMIC_FIELD_LENGTH_SIZE_BYTES > len(bet_bytes):
        raise ValueError("Data is too short to contain field length")

    field_length , position = _get_int_from_field(bet_bytes, position, DYNAMIC_FIELD_LENGTH_SIZE_BYTES)

    if position + field_length > len(bet_bytes):
        raise ValueError("Data is too short to contain the dynamic field")

    field , position = _get_str_from_field(bet_bytes, position, field_length)

    return field, position


def _get_int_from_field(bet_bytes: bytes, position: int, length: int):
    return int.from_bytes(bet_bytes[position:position + length], byteorder='big'), position + length

def _get_str_from_field(bet_bytes: bytes, position: int, length: int):
    return bet_bytes[position:position + length].decode('utf-8') , position + length