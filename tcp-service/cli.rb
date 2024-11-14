require 'socket'
require 'thread'

class NetworkClient
  def initialize(host, port)
    @host = host
    @port = port
    @socket = nil
    @running = false
    @username = nil
  end

  def connect
    @socket = TCPSocket.new(@host, @port)
    puts "Connected to server at #{@host}:#{@port}"
    @running = true
    Thread.new { receive_messages }
    true
  rescue StandardError => e
    puts "Failed to connect: #{e}"
    false
  end

  def disconnect
    if @socket
      @running = false
      @socket.close
      puts "Disconnected from server"
    end
  end

  def send_message(message)
    return unless @socket

    begin
      @socket.puts(message)
    rescue StandardError => e
      puts "Failed to send message: #{e}"
    end
  end

  def receive_messages
    while @running
      begin
        data = @socket.gets&.chomp
        if data
          puts "Received: #{data}"
        else
          puts "Server closed the connection"
          disconnect
          break
        end
      rescue StandardError => e
        puts "Error receiving message: #{e}" if @running
        break
      end
    end
  end

  def login_as_admin(username, password)
    send_message("@login_as_admin #{username} #{password}")
  end

  def broadcast_message(message)
    send_message("@broadcast #{message}")
  end

  def execute_server_command(command)
    send_message("@server-command #{command}")
  end

  def exit
    send_message("@exit")
    disconnect
  end
end

def main
  client = NetworkClient.new('localhost', 8080)
  if client.connect
    loop do
      begin
        print "Enter a message (or 'exit' to quit): "
        message = gets.chomp
        if message.downcase == 'exit'
          client.exit
          break
        elsif message.start_with?('@')
          case message
          when /^@login_as_admin/
            parts = message.split
            if parts.length == 3
              client.login_as_admin(parts[1], parts[2])
            else
              puts "Usage: @login_as_admin username password"
            end
          when /^@broadcast/
            client.broadcast_message(message[10..-1])
          when /^@server-command/
            client.execute_server_command(message[15..-1])
          else
            client.send_message(message)
          end
        else
          client.send_message(message)
        end
      rescue Interrupt
        # Handle Ctrl+C without quitting
        puts "\n(KeyboardInterrupt detected, type 'exit' to quit the program.)"
      end
    end
  end
end

if __FILE__ == $0
  main
end
