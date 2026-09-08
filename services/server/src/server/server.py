import socket
import threading
import signal
import logger
import protocol
from lottery import Lottery


ACTION_ACEPT_CONNECTION = "accept-connection"
ACTION_HANDLE_CLIENT = "handle-client"
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

    def _handle_sigterm(self, signum, frame):
        logger.info("signal-received", logger.LogResult.in_progress, "signal", signum)

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
        for thread in self.client_threads:
            thread.join(timeout=THREAD_JOIN_TIMEOUT_SECONDS)
        if thread.is_alive():
            logger.warning("thread-shutdown", logger.LogResult.fail, "thread", thread.name)
                

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
        self._register_client_socket(client_socket)
        try:
            with client_socket:
                logger.info(ACTION_HANDLE_CLIENT, logger.LogResult.in_progress)

                first_batch = self._store_bets(client_socket)
                if not self.running:
                    return
                
                agency_id = self._agency_id_from(first_batch)

                # block thread until the quorum is done
                quorum_reached = self._wait_for_quorum()
                if not quorum_reached:
                    return
                    
                winners_bets = self._winners_for_agency(agency_id)

                self._send_winners_bets(client_socket, winners_bets)
                self._send_end_of_sending(client_socket)

                logger.info(ACTION_HANDLE_CLIENT, logger.LogResult.success)    

        except Exception as error:
                if self.running:
                    logger.error(ACTION_HANDLE_CLIENT, logger.LogResult.fail,"err", error)
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

    def _store_bets(self, client_socket):
        """
        Receives bets from the client and stores them in the lottery. 
        Returns the first batch of bets received.
        """
        first_batch = []
        try:
            for i, bets_batch in enumerate(protocol.receive_bets(client_socket)):
                self.lottery.store_bets(bets_batch)
                if i == 0:
                    first_batch = bets_batch
                protocol.send_batch_ack(client_socket, success=True)

        except Exception as error:
            if self.running:
                protocol.send_batch_ack(client_socket, success=False)
                logger.error(ACTION_HANDLE_CLIENT, logger.LogResult.fail, "err", error)
                raise error

        return first_batch

    def _agency_id_from(self, bets_batch):
        return bets_batch[0].agency_id if bets_batch else None

    def _wait_for_quorum(self):
        with self.quorum_condition:
            self.agencies_finished += 1
            logger.info("quorum-check",logger.LogResult.in_progress, "agencies_finished", self.agencies_finished)

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

    def _send_winners_bets(self, client_socket, winners_bets):
        for winner_bet in winners_bets:
            protocol.send_winner_bet(client_socket, winner_bet)

    def _send_end_of_sending(self, client_socket):
        protocol.send_end(client_socket)