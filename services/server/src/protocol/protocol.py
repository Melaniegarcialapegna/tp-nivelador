import safe_socket
import logger  
from collections.abc import Iterator
from lottery import Lottery
from lottery import Bet 

#Protocol
MESSAGE_TYPE_BET = 0 #context of bets
MESSAGE_TYPE_END = 1 #finish of sending bets
MESSAGE_TYPE_ACK_OK = 2 #ack of reception without error
MESSAGE_TYPE_ACK_FAIL = 3 #ack of reception with error
#-------------------------------------------------------

TYPE_MESSAGE_SIZE_BYTES = 1
ACK_MESSAGE_SIZE_BYTES = 1

LENGTH_FIELD_SIZE_BYTES = 4

HEADER_SIZE_BYTES = TYPE_MESSAGE_SIZE_BYTES + LENGTH_FIELD_SIZE_BYTES

EMPTY_MESSAGE = 0

## Serializarion 

FIXED_FIELDS_SIZE_BYTES = 4 
BIRTHDATE_LENGTH = 10 #YYYY-MM-DD

DYNAMIC_FIELD_LENGTH_SIZE_BYTES = 2 

AMOUNT_BYTES_CONST = FIXED_FIELDS_SIZE_BYTES * 3 + BIRTHDATE_LENGTH  # agency_id + document + number + birthdate

class Protocol:
    def __init__(self, socket):
        self.socket = socket

    def receive_bets(self) -> Iterator[list[Bet]]:
        action = "receive-bets"
        try:  
            header_buffer = self._receive_header()

            while header_buffer[0] == MESSAGE_TYPE_BET: 

                lenght_batch = self._length_from_header(header_buffer)
                batch_bytes = safe_socket.recv_all(self.socket, lenght_batch) 

                #Return the bets in a batch
                yield self._parse_batch(batch_bytes)

                header_buffer = self._receive_header()

            self._check_end_of_bets(action,header_buffer)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, "exception", str(e))
            raise e

    def send_batch_ack(self, success):
        message_type = MESSAGE_TYPE_ACK_OK if success else MESSAGE_TYPE_ACK_FAIL
        ack_message = (message_type).to_bytes(ACK_MESSAGE_SIZE_BYTES, byteorder='big')
        safe_socket.send_all(self.socket, ack_message)

    def send_winner_bet(self, winner_bet):
        bet_bytes = self._serialize_bet(winner_bet)
        header = self._create_header_for_bet(len(bet_bytes))

        bet_message = header + bet_bytes
        safe_socket.send_all(self.socket, bet_message)

    def send_end(self):
        message_end = bytes([MESSAGE_TYPE_END]) + (EMPTY_MESSAGE).to_bytes(LENGTH_FIELD_SIZE_BYTES, byteorder='big')
        safe_socket.send_all(self.socket, message_end)

    #privates

    def _receive_header(self):
        return safe_socket.recv_all(self.socket, HEADER_SIZE_BYTES)  

    def _length_from_header(self, header_buffer):
        return int.from_bytes(header_buffer[TYPE_MESSAGE_SIZE_BYTES:HEADER_SIZE_BYTES], byteorder='big')

    def _parse_batch(self, bet_bytes):
        bets = []
        for bet_bytes in self._separate_bets_from(bet_bytes):
            bet = self._deserialize_bet(bet_bytes)
            bets.append(bet)
        return bets

    def _separate_bets_from(self,batch_bytes: bytes) -> Iterator[bytes]:
        position = 0
        while position < len(batch_bytes):

            if position + LENGTH_FIELD_SIZE_BYTES > len(batch_bytes):
                raise ValueError("Data too short to know the length of the next bet")
            
            length_bet = int.from_bytes(batch_bytes[position:position + LENGTH_FIELD_SIZE_BYTES], byteorder='big')

            position += LENGTH_FIELD_SIZE_BYTES

            if position + length_bet > len(batch_bytes):
                raise ValueError("Data too short for a bet")
            
            bet_bytes = batch_bytes[position:position + length_bet]

            yield bet_bytes #return to be processed 

            position += length_bet  


    def _check_end_of_bets(self,action,header_buffer):
        if header_buffer[0] != MESSAGE_TYPE_END:
            logger.error(action, logger.LogResult.fail, "unexpected-message-type")
            raise ValueError("Unexpected message type received")

    def _create_header_for_bet(self,length_bet_bytes: int) -> bytes:
        message_type = bytes([MESSAGE_TYPE_BET])

        length_bytes = length_bet_bytes.to_bytes(LENGTH_FIELD_SIZE_BYTES, byteorder='big')
        return message_type + length_bytes

    ### ----- Serializarion

    # Serializes a bet to a byte array
    # For the fields that are dinamic in size, will be used a separator to know how many bytes to read for each field 
    # long_dinamic_field_i | dinamic_field_i |
    def _serialize_bet(self,bet: Bet) -> bytes:
        #put fixed fields 
        bet_bytes = self._get_field_bytes(bet.agency_id)
        bet_bytes += self._get_field_bytes(bet.document)
        bet_bytes += self._get_field_bytes(bet.number)
        bet_bytes += str(bet.birthdate).encode('utf-8')

        #put dinamic fields
        bet_bytes += self._get_dynamic_field(bet.first_name)
        bet_bytes += self._get_dynamic_field(bet.last_name)

        return bet_bytes

    def _get_field_bytes(self, field) -> bytes:
        return field.to_bytes(FIXED_FIELDS_SIZE_BYTES, byteorder='big')

    def _get_dynamic_field(self, field: str) -> bytes:
        #convert field to bytes
        field_bytes = field.encode('utf-8')

        #length of the field in bytes
        field_length = len(field_bytes)
        field_length_bytes = field_length.to_bytes(DYNAMIC_FIELD_LENGTH_SIZE_BYTES, byteorder='big')

        return field_length_bytes + field_bytes #header of the bet

    #--

    #Deserializes a byte array to a bet
    def _deserialize_bet(self, bet_bytes: bytes) -> Bet:
        if len(bet_bytes) < AMOUNT_BYTES_CONST:
            raise ValueError("Data too short to deserialize a bet")

        position = 0

        agency_id , position = self._get_int_from_field(bet_bytes, position, FIXED_FIELDS_SIZE_BYTES)
        document , position = self._get_int_from_field(bet_bytes, position, FIXED_FIELDS_SIZE_BYTES)
        number , position = self._get_int_from_field(bet_bytes, position, FIXED_FIELDS_SIZE_BYTES)

        birthdate , position = self._get_str_from_field(bet_bytes, position, BIRTHDATE_LENGTH)

        first_name, position = self._read_dynamic_field(bet_bytes, position)

        last_name, _ = self._read_dynamic_field(bet_bytes, position)

        return Bet(
            agency_id=agency_id,
            first_name=first_name,
            last_name=last_name,
            document=document,
            birthdate=birthdate,
            number=number)

    def _read_dynamic_field(self, bet_bytes: bytes, position: int):
        if position + DYNAMIC_FIELD_LENGTH_SIZE_BYTES > len(bet_bytes):
            raise ValueError("Data is too short to contain field length")

        field_length , position = self._get_int_from_field(bet_bytes, position, DYNAMIC_FIELD_LENGTH_SIZE_BYTES)

        if position + field_length > len(bet_bytes):
            raise ValueError("Data is too short to contain the dynamic field")

        field , position = self._get_str_from_field(bet_bytes, position, field_length)

        return field, position



    def _get_int_from_field(self, bet_bytes: bytes, position: int, length: int):
        return int.from_bytes(bet_bytes[position:position + length], byteorder='big'), position + length

    def _get_str_from_field(self, bet_bytes: bytes, position: int, length: int):
        return bet_bytes[position:position + length].decode('utf-8') , position + length