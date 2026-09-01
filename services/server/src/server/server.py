import socket
import logger
import protocol
from lottery import Lottery


ACTION_ACEPT_CONNECTION = "accept-connection"
ACTION_HANDLE_CLIENT = "handle-client"

class Server:
    def __init__(self, server_host: str, server_port: int,storage_path: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(storage_path) # shared with all the clients

    def run(self):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                client_socket = self._accept_connection(server_socket)
                self._handle_client(client_socket)

    def _accept_connection(self, server_socket):
        logger.info(ACTION_ACEPT_CONNECTION, logger.LogResult.in_progress)

        try:
            client_socket, _ = server_socket.accept()
        except Exception as error:
            logger.error(ACTION_ACEPT_CONNECTION, logger.LogResult.fail, "err", error)
            raise error
        
        logger.info(ACTION_ACEPT_CONNECTION, logger.LogResult.success)
        return client_socket
        

    def _handle_client(self, client_socket):
        try:
            logger.info(ACTION_HANDLE_CLIENT, logger.LogResult.in_progress)

            first_batch = self._store_bets(client_socket)
            agency_id = self._agency_id_from(first_batch)
                
            winners_bets = self._winners_for_agency(agency_id)

            self._send_winners_bets(client_socket, winners_bets)
            self._send_end_of_sending(client_socket)

            logger.info(ACTION_HANDLE_CLIENT, logger.LogResult.success)    

        except Exception as error:
            logger.error(ACTION_HANDLE_CLIENT, logger.LogResult.fail,"err", error)
            raise error

    def _store_bets(self, client_socket):
        """
        Receives bets from the client and stores them in the lottery. 
        Returns the first batch of bets received.
        """
        try:
            for i, bets_batch in enumerate(protocol.receive_bets(client_socket)):
                self.lottery.store_bets(bets_batch)
                if i == 0:
                    first_batch = bets_batch
                protocol.send_batch_ack(client_socket, success=True)

        except Exception as error:
            protocol.send_batch_ack(client_socket, success=False)
            logger.error(ACTION_HANDLE_CLIENT, logger.LogResult.fail, "err", error)
            raise error

        return first_batch

    def _agency_id_from(self, bets_batch):
        return bets_batch[0].agency_id if bets_batch else None

    def _winners_for_agency(self, agency_id: str):
        winners_bets = []

        for bet in self.lottery.load_bets():
            if bet.agency_id == agency_id and self.lottery.has_won(bet):
                winners_bets.append(bet)

        return winners_bets

    def _send_winners_bets(self, client_socket, winners_bets):
        for winner_bet in winners_bets:
            protocol.send_winner_bet(client_socket, winner_bet)

    def _send_end_of_sending(self, client_socket):
        protocol.send_end(client_socket)