import socket


def send_all(socket: socket.socket, bytes):
    total_sent = 0
    while total_sent < len(bytes): #loop continue until all bytes are sent
        size_sent = socket.send(bytes[total_sent:])
        if size_sent == 0: 
            raise ConnectionError("Connection closed before sending all data")
        total_sent += size_sent
    return total_sent

def recv_all(socket: socket.socket, size):
    data = b""
    while len(data) < size: #loop continue until all bytes are read
        data_read = socket.recv(size-len(data))
        if data_read == b"": 
            raise ConnectionError("Connection closed before reading all data")
        data += data_read
    return data