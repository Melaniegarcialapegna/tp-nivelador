import safe_socket
import logger  
from src_frozen import Bet 
import protocol

MESSAGE_TYPE_BET = 0 # hay una apuesta/ganador
MESSAGE_TYPE_END = 1 #termino de enviar apuestas/ganadores


MESSAGE_TYPE_BET = 0
HEADER_AMOUNT = 5
TYPE_AMOUNT = 1
AMOUNT_BYTES_UINT32 = 4 #TODO : cambiar

def receive_bets(socket) -> list:
    #action = "receive-bets" VER 
    #misma idea de ReceiveWinners
    try: 
        bets = []
        header_buffer = safe_socket.recv_all(socket, HEADER_AMOUNT) #recibo header
        while header_buffer[0] == MESSAGE_TYPE_BET: 
            lenght_payload = int.from_bytes(header_buffer[TYPE_AMOUNT:HEADER_AMOUNT], byteorder='big') 

            bet_bytes = safe_socket.recv_all(socket, lenght_payload) 

            #convierto a bet y agrego a lista de bets
            bet = protocol.deserialize_bet(bet_bytes) 

            bets.append(bet) 

            header_buffer = safe_socket.recv_all(socket, HEADER_AMOUNT) 

        if header_buffer[0] != MESSAGE_TYPE_END:
            logger.error("receive-bets", logger.LogResult.fail, "unexpected-message-type")
            raise Exception("Unexpected message type received") 

        return bets
    except Exception as e:
        logger.error("receive-bets", logger.LogResult.fail, "exception", str(e))
        raise e

#def send_winner_bet(socket, winner_bet):

def send_end(socket):
    message_end = bytes([MESSAGE_TYPE_END]) + (0).to_bytes(AMOUNT_BYTES_UINT32, byteorder='big')
    safe_socket.send_all(socket, message_end)
