import socket
import logger
import protocol
from lottery import Lottery

class Server:
    def __init__(self, server_host: str, server_port: int,storage_path: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(storage_path) #compartida para todos los clientes 

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)

    def _handle_client(self, client_socket):
        action = "handle-client"
        try:
            logger.info(action, logger.LogResult.in_progress)

            primera = True
            bet_modelo = None
            # va recibiendo y persistiendo las bets
            for bets_batch in protocol.receive_bets(client_socket):
                self.lottery.store_bets(bets_batch)
                if primera:
                    bet_modelo = bets_batch[0] if bets_batch else None
                    primera = False
            #calcula ganadores de esta agencia y los envia
            agency_id = bet_modelo.agency_id if bet_modelo else None

            winners_bets = []

            for bet in self.lottery.load_bets():
                if bet.agency_id == agency_id and self.lottery.has_won(bet):
                    winners_bets.append(bet)

            for winner_bet in winners_bets:
                protocol.send_winner_bet(client_socket, winner_bet) 

            #aviso que termine
            protocol.send_end(client_socket)

            logger.info(action, logger.LogResult.success)    

        except Exception as e:
            logger.error(
                action, logger.LogResult.fail
            )
            raise e