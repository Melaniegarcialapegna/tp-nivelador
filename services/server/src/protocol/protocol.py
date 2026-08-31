import safe_socket
import logger  
from .serialization import deserialize_bet, serialize_bet
from collections.abc import Iterator
from lottery import Lottery

MESSAGE_TYPE_BET = 0 # hay una apuesta/ganador
MESSAGE_TYPE_END = 1 #termino de enviar apuestas/ganadores
MESSAGE_TYPE_ACK_OK = 2 #ack de recepcion correcta
MESSAGE_TYPE_ACK_FAIL = 3 #ack de recepcion con error


MESSAGE_TYPE_BET = 0
HEADER_AMOUNT = 5
TYPE_AMOUNT = 1
AMOUNT_BYTES_UINT32 = 4 #TODO : cambiar
AMOUNT_BYTES_UINT16 = 2 #TODO : cambiar
AMOUNT_BYTES_UINT8 = 1 #TODO : cambiar

def receive_bets(socket) -> Iterator[list[Bet]]:
    #action = "receive-bets" VER 
    #misma idea de ReceiveWinners
    try: 
        
        header_buffer = safe_socket.recv_all(socket, HEADER_AMOUNT) #recibo header
        while header_buffer[0] == MESSAGE_TYPE_BET: 
            lenght_payload = int.from_bytes(header_buffer[TYPE_AMOUNT:HEADER_AMOUNT], byteorder='big')

            batch_bytes = safe_socket.recv_all(socket, lenght_payload) 
            batch_bets = []
            for bet_bytes in separate_bets(batch_bytes):
                #convierto a bet 
                bet = deserialize_bet(bet_bytes) 
                bet = deserialize_bet(bet_bytes) 
                batch_bets.append(bet)
            #entrego a proxima capa
            yield batch_bets

            header_buffer = safe_socket.recv_all(socket, HEADER_AMOUNT) 

        if header_buffer[0] != MESSAGE_TYPE_END:
            logger.error("receive-bets", logger.LogResult.fail, "unexpected-message-type")
            raise Exception("Unexpected message type received") 

    except Exception as e:
        logger.error("receive-bets", logger.LogResult.fail, "exception", str(e))
        raise e

def separate_bets(batch_bytes: bytes) -> Iterator[bytes]:
    position = 0

    while position < len(batch_bytes):
        if position + AMOUNT_BYTES_UINT32 > len(batch_bytes):
            raise ValueError("Data too short to know the length of the next bet")
        
        length_bet = int.from_bytes(batch_bytes[position:position + AMOUNT_BYTES_UINT32], byteorder='big')

        position += AMOUNT_BYTES_UINT32

        if position + length_bet > len(batch_bytes):
            raise ValueError("Data too short for a bet")
        
        bet_bytes = batch_bytes[position:position + length_bet]
        yield bet_bytes #retorno para que vaya siendo procesada

        position += length_bet    


def send_batch_ack(socket,success):
    message_type = MESSAGE_TYPE_ACK_OK if success else MESSAGE_TYPE_ACK_FAIL
    ack_message = (message_type).to_bytes(AMOUNT_BYTES_UINT8, byteorder='big')
    safe_socket.send_all(socket, ack_message)

def send_winner_bet(socket, winner_bet):
    payloadBet = serialize_bet(winner_bet)
    header = createHeader(len(payloadBet))

    betMessage = header + payloadBet
    safe_socket.send_all(socket, betMessage)

def createHeader(payload_length: int) -> bytes:
    message_type = bytes([MESSAGE_TYPE_BET])
    length_bytes = payload_length.to_bytes(AMOUNT_BYTES_UINT32, byteorder='big')
    return message_type + length_bytes

def send_end(socket):
    message_end = bytes([MESSAGE_TYPE_END]) + (0).to_bytes(AMOUNT_BYTES_UINT32, byteorder='big')
    safe_socket.send_all(socket, message_end)
