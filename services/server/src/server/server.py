import socket
import threading
import signal
import logger
import protocol
from lottery import Lottery

FIRST_BET = 0
FIRST_BATCH = 0
THREAD_JOIN_TIMEOUT_SECONDS = 4

class Server:
    def __init__(self, server_host: str, server_port: int,storage_path: str, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(storage_path) # shared with all the clients

        self.file_lock = threading.Lock()  # Lock for file access
        self.agency_quorum_min = agency_quorum_min
        self.agencies_finished = 0
        self.quorum_condition = threading.Condition()

        self.running = True
        self.server_socket = None
        self.client_threads = []
        self.active_clients = []
        self.clients_lock = threading.Lock()

        signal.signal(signal.SIGTERM, self._handle_sigterm)
        signal.signal(signal.SIGINT, self._handle_sigterm)

    def _handle_sigterm(self, signum, frame):
        action = "sigterm-received"
        logger.info(action, logger.LogResult.in_progress, "signal", signum)

        self.running = False

        # wake up threads waiting for quorum
        with self.quorum_condition:
            self.quorum_condition.notify_all()

        # desblock the accept of principal loop
        if self.server_socket:
            try:
                self.server_socket.close()
            except OSError:
                pass

        # desblock any receive for a client thread
        with self.clients_lock:
            for client_socket in self.active_clients:
                try:
                    client_socket.shutdown(socket.SHUT_RDWR)
                except OSError:
                    pass
                try:
                    client_socket.close()
                except OSError:
                    pass

    def run(self):
        self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            self.server_socket.bind((self.server_host, self.server_port))
            self.server_socket.listen()
            while self.running:
                try:
                    client_socket = self._accept_connection(self.server_socket)
                except OSError:
                    break # Socket closed in _handle_sigterm

                # For each client a new thread and continue accepting connections 
                thread = threading.Thread( target=self._handle_client, args=(client_socket,) )
                self.client_threads.append(thread)
                thread.start()
        finally:
            try:
                self.server_socket.close()
            except OSError:
                pass
            self._wait_for_client_threds()

    def _wait_for_client_threds(self):
        action = "thread-shutdown"
        logger.info(action, logger.LogResult.in_progress)
        for thread in self.client_threads:
            thread.join(timeout=THREAD_JOIN_TIMEOUT_SECONDS)
            if thread.is_alive():
                logger.error(action, logger.LogResult.fail, "thread", thread.name)
                

    def _accept_connection(self, server_socket):
        action = "accept-connection"
        logger.info(action, logger.LogResult.in_progress)

        try:
            client_socket, _ = server_socket.accept()
        except Exception as error:
            logger.error(action, logger.LogResult.fail, "err", error)
            raise error
        
        logger.info(action, logger.LogResult.success)
        return client_socket

    def _handle_client(self, client_socket):
        action = "handle-client"
        self._register_client_socket(client_socket)
        try:
            with client_socket:
                logger.info(action, logger.LogResult.in_progress)

                proto = protocol.Protocol(client_socket)

                first_batch = self._store_bets(action,proto)
                if not self.running:
                    return
                
                agency_id = self._agency_id_from(first_batch)

                # block thread until the quorum is done
                quorum_reached = self._wait_for_quorum()
                if not quorum_reached:
                    return
                    
                winners_bets = self._winners_for_agency(agency_id)

                self._send_winners_bets(proto, winners_bets)
                self._send_end_of_sending(proto)

                logger.info(action, logger.LogResult.success)    

        except Exception as error:
                if self.running:
                    logger.error(action, logger.LogResult.fail,"err", error)
                    raise error
        finally:
            self._unregister_client_socket(client_socket)

    def _register_client_socket(self, sock):
        with self.clients_lock:
            self.active_clients.append(sock)

    def _unregister_client_socket(self, sock):
        with self.clients_lock:
            if sock in self.active_clients:
                self.active_clients.remove(sock)

    def _store_bets(self, action, protocol):
        """
        Receives bets from the client and stores them in the lottery. 
        Returns the first batch of bets received.
        """
        first_batch = []
        try:
            for i, bets_batch in enumerate(protocol.receive_bets()):
                with self.file_lock:
                    self.lottery.store_bets(bets_batch)
                    if i == FIRST_BATCH:
                        first_batch = bets_batch
                    protocol.send_batch_ack(success=True)

        except Exception as error:
            if self.running:
                protocol.send_batch_ack(success=False)
                logger.error(action, logger.LogResult.fail, "err", error)
                raise error

        return first_batch

    def _agency_id_from(self, bets_batch):
        return bets_batch[FIRST_BET].agency_id if bets_batch else None

    def _wait_for_quorum(self):
        action = "quorum-check"
        with self.quorum_condition:
            self.agencies_finished += 1
            logger.info(action,logger.LogResult.in_progress, "agencies_finished", self.agencies_finished)

            if self.agencies_finished >= self.agency_quorum_min:
                self.quorum_condition.notify_all()
            else:
                self.quorum_condition.wait_for(lambda: self.agencies_finished >= self.agency_quorum_min or not self.running)

            return self.agencies_finished >= self.agency_quorum_min

    def _winners_for_agency(self, agency_id: str):
        winners_bets = []

        with self.file_lock:
            for bet in self.lottery.load_bets():
                if bet.agency_id == agency_id and self.lottery.has_won(bet):
                    winners_bets.append(bet)

        return winners_bets

    def _send_winners_bets(self, protocol, winners_bets):
        for winner_bet in winners_bets:
            protocol.send_winner_bet(winner_bet)

    def _send_end_of_sending(self, protocol):
        protocol.send_end()