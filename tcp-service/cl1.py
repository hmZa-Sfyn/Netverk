import socket
import threading
import time

class NetworkClient:
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.socket = None
        self.running = False
        self.username = None

    def connect(self):
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            self.socket.connect((self.host, self.port))
            print(f"Connected to server at {self.host}:{self.port}")
            self.running = True
            threading.Thread(target=self.receive_messages, daemon=True).start()
            return True
        except Exception as e:
            print(f"Failed to connect: {e}")
            return False

    def disconnect(self):
        if self.socket:
            self.running = False
            self.socket.close()
            print("Disconnected from server")

    def send_message(self, message):
        if self.socket:
            try:
                self.socket.sendall(f"{message}\n".encode())
            except Exception as e:
                print(f"Failed to send message: {e}")

    def receive_messages(self):
        while self.running:
            try:
                data = self.socket.recv(1024).decode().strip()
                if data:
                    print(f"Received: {data}")
                else:
                    print("Server closed the connection")
                    self.disconnect()
                    break
            except Exception as e:
                if self.running:
                    print(f"Error receiving message: {e}")
                break

    def login_as_admin(self, username, password, session_id):
        self.send_message(f"@login_as_admin {username} {password} {session_id}")

    def broadcast_message(self, message):
        self.send_message(f"@broadcast {message}")

    def execute_server_command(self, command):
        self.send_message(f"@server-command {command}")

    def exit(self):
        self.send_message("@exit")
        self.disconnect()

def main():
    client = NetworkClient('localhost', 8080)
    if client.connect():
        try:
            while True:
                try:
                    message = input("Enter a message (or 'exit' to quit): ")
                    if message.lower() == 'exit':
                        client.exit()
                        break
                    elif message.startswith('@'):
                        if message.startswith('@login_as_admin'):
                            parts = message.split()
                            if len(parts) == 4:
                                client.login_as_admin(parts[1], parts[2], parts[3])
                            else:
                                print("Usage: @login_as_admin username password session_id")
                        elif message.startswith('@broadcast'):
                            client.broadcast_message(message[10:])
                        elif message.startswith('@server-command'):
                            client.execute_server_command(message[15:])
                        else:
                            client.send_message(message)
                    else:
                        client.send_message(message)
                except KeyboardInterrupt:
                    # Handle Ctrl+C without quitting
                    print("\n(KeyboardInterrupt detected, type 'exit' to quit the program.)")
                    continue  # Simply continue the loop

        finally:
            client.disconnect()

if __name__ == "__main__":
    main()
