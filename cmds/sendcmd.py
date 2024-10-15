import socket
import time
import sys


def netcat(hostname, port, content):
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect((hostname, port))
    s.sendall(content)
    time.sleep(0.5)
    s.shutdown(socket.SHUT_WR)
    while 1:
        data = s.recv(1024)
        if len(data) == 0:
            break
        print(data[0:4].decode('utf8') + data[6:].decode('utf8'))
    s.close()


def main():
    args = sys.argv[1:]

    if len(args) < 4:
        return

    hostname = args[0]
    port = int(args[1])
    header = args[2]
    content = args[3]

    encoded_content_size = (len(content)).to_bytes(
        2,
        byteorder='big'
    )

    netcat(
        hostname,
        port,
        header.encode() + encoded_content_size + content.encode()
    )


if __name__ == "__main__":
    main()
