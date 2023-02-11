import socket
from cache import cache


def start_server():
    # Create a TCP/IP socket
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    # Bind the socket to a specific address and port
    server_address = ('localhost', 6379)
    server_socket.bind(server_address)

    # Listen for incoming connections
    server_socket.listen(1)

    while True:
        # Wait for a connection
        print('waiting for a connection')
        client_socket, client_address = server_socket.accept()
        print('connection from', client_address)

        # Receive data from the client
        data = client_socket.recv(1024).decode()
        data = data.strip().split()

        if not data:
            response = 'Invalid command'
        else:
            if data[0] == 'GET':
                # Retrieve value from cache
                key = data[1]
                if key in cache:
                    response = cache[key]
                else:
                    response = 'Key not found'

            elif data[0] == 'SET':
                # Store value in cache
                key = data[1]
                value = data[2]
                cache[key] = value
                response = 'Value stored'

            else:
                response = 'Invalid command'

        # Send response back to the client
        client_socket.sendall(response.encode())

        # Clean up the connection
        client_socket.close()
