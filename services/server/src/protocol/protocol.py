import safe_socket
import logger  
from .serialization import deserialize_bet, serialize_bet
from collections.abc import Iterator
from lottery import Lottery

#Protocol
MESSAGE_TYPE_BET = 0 #context of bets
MESSAGE_TYPE_END = 1 #finish of sending bets
MESSAGE_TYPE_ACK_OK = 2 #ack of reception without error
MESSAGE_TYPE_ACK_FAIL = 3 #ack of reception with error
#-------------------------------------------------------

ACTION_RECEIVE_BETS = "receive-bets"

TYPE_MESSAGE_SIZE_BYTES = 1
ACK_MESSAGE_SIZE_BYTES = 1

LENGTH_FIELD_SIZE_BYTES = 4

HEADER_SIZE_BYTES = TYPE_MESSAGE_SIZE_BYTES + LENGTH_FIELD_SIZE_BYTES

EMPTY_MESSAGE = 0


def receive_bets(socket) -> Iterator[list[Bet]]:
    try:  
        header_buffer = _receive_header(socket)

        while _still_receive_bets(header_buffer): 

            lenght_batch = _length_from_header(header_buffer)
            batch_bytes = safe_socket.recv_all(socket, lenght_batch) 

            #Return the bets
            yield _parse_batch(batch_bytes)

            header_buffer = _receive_header(socket)

        _check_end_of_bets(header_buffer)

    except Exception as e:
        logger.error(ACTION_RECEIVE_BETS, logger.LogResult.fail, "exception", str(e))
        raise e

def _receive_header(socket):
    return safe_socket.recv_all(socket, HEADER_SIZE_BYTES)  

def _still_receive_bets(header_buffer):
    return header_buffer[0] == MESSAGE_TYPE_BET

def _length_from_header(header_buffer):
    return int.from_bytes(header_buffer[TYPE_MESSAGE_SIZE_BYTES:HEADER_SIZE_BYTES], byteorder='big')

def _parse_batch(bet_bytes):
    bets = []
    for bet_bytes in _separate_bets_from(bet_bytes):
        bet = deserialize_bet(bet_bytes)
        bets.append(bet)
    return bets

def _separate_bets_from(batch_bytes: bytes) -> Iterator[bytes]:
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


def _check_end_of_bets(header_buffer):
    if header_buffer[0] != MESSAGE_TYPE_END:
        logger.error(ACTION_RECEIVE_BETS, logger.LogResult.fail, "unexpected-message-type")
        raise ValueError("Unexpected message type received")

#--

def send_batch_ack(socket,success):
    message_type = MESSAGE_TYPE_ACK_OK if success else MESSAGE_TYPE_ACK_FAIL
    ack_message = (message_type).to_bytes(ACK_MESSAGE_SIZE_BYTES, byteorder='big')
    safe_socket.send_all(socket, ack_message)

#--

def send_winner_bet(socket, winner_bet):
    bet_bytes = serialize_bet(winner_bet)
    header = _create_header_for_bet(len(bet_bytes))

    bet_message = header + bet_bytes
    safe_socket.send_all(socket, bet_message)

def _create_header_for_bet(length_bet_bytes: int) -> bytes:
    message_type = bytes([MESSAGE_TYPE_BET])

    length_bytes = length_bet_bytes.to_bytes(LENGTH_FIELD_SIZE_BYTES, byteorder='big')
    return message_type + length_bytes

#--

def send_end(socket):
    message_end = bytes([MESSAGE_TYPE_END]) + (EMPTY_MESSAGE).to_bytes(LENGTH_FIELD_SIZE_BYTES, byteorder='big')
    safe_socket.send_all(socket, message_end)
