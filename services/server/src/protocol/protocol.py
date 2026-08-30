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