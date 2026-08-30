import socket
import logger
import protocol
from src_frozen import Lottery

_ECHO_SERVER_MESSAGE_SIZE = 1024


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

            # recibe bets y las persiste
            bets = protocol.receive_bets(client_socket)
            self.lottery.store_bets(bets)

            #calcula ganadores de esta agencia y los envia
            agency_id = bets[0].agency_id if bets else None

            winners_bets = []

            for bet in self.lottery.load_bets():
                if bet.agency_id == agency_id and self.lottery.is_winner(bet):
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