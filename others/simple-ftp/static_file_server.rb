require "net/simple_server"

server = Net::SimpleServer.new("3000")
server.file_root = "./../../samples"

server.start