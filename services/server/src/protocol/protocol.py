import safe_socket
import logger  
from src_frozen import Bet 

MESSAGE_TYPE_BET = 0
HEADER_AMOUNT = 5
TYPE_AMOUNT = 1

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
            bet = deserialize_bet(bet_bytes) 

            bets.append(bet) 

            header_buffer = safe_socket.recv_all(socket, HEADER_AMOUNT) 

        if header_buffer[0] != 1:
            logger.error("receive-bets", logger.LogResult.fail, "unexpected-message-type")
            raise Exception("Unexpected message type received") 

        return bets
    except Exception as e:
        logger.error("receive-bets", logger.LogResult.fail, "exception", str(e))
        raise e

#def send_winner_bet(socket, winner_bet):

#def send_end(socket):


            # while True:

            #     client_message = safe_socket.recv_all(
            #         client_socket, _ECHO_SERVER_MESSAGE_SIZE
            #     )
            #     if not client_message:
            #         logger.info(
            #             action,
            #             logger.LogResult.success,
            #             "messages-amount",
            #             message_amount,
            #         )
            #         return
            #     safe_socket.send_all(client_socket, client_message)