require 'socket'
require 'securerandom'
require 'json'
require 'fileutils'

module MrRobotNet
  class NetworkServer
    def initialize(port)
      @server = TCPServer.new(port)
      @connected_clients = {}
      @current_session_id = generate_session_id
      @admin_session_id = generate_session_id
      @admin_user = nil
      @message_counters = {}
      @last_message_times = {}
      initialize_directories
    end

    def initialize_directories
      dirs = [
        "./#{@current_session_id}",
        "./#{@current_session_id}/nodes/client/all",
        "./#{@current_session_id}/nodes/client/banned",
        "./#{@current_session_id}/transactions/#{generate_session_id}"
      ]
      dirs.each { |dir| FileUtils.mkdir_p(dir) }
    end

    def start
      puts "Server started. Listening on port #{@server.addr[1]}"
      puts "Admin session ID: #{@admin_session_id}"

      loop do
        client = @server.accept
        Thread.new { handle_client(client) }
      end
    end

    private

    def handle_client(client)
      client_id = SecureRandom.uuid
      user_node = UserNode.new(client_id, NetworkPermissions.new(send: true, receive: true, broadcast:true))
      @connected_clients[client_id] = user_node
      @message_counters[client_id] = 0
      @last_message_times[client_id] = Time.now

      log_client_connection(client_id, client.peeraddr[3])

      begin
        loop do
          message = client.gets&.chomp
          break if message.nil?

          if rate_limited?(client_id)
            ban_user(client_id, "Rate limit exceeded")
            break
          end

          process_message(client_id, message, client)
        end
      ensure
        log_client_disconnection(client_id)
        @connected_clients.delete(client_id)
        @message_counters.delete(client_id)
        @last_message_times.delete(client_id)
        client.close
      end
    end

    def rate_limited?(client_id)
      now = Time.now
      if (now - @last_message_times[client_id]) <= 4
        @message_counters[client_id] += 1
        return true if @message_counters[client_id] > 49
      else
        @message_counters[client_id] = 1
        @last_message_times[client_id] = now
      end
      false
    end

    def process_message(client_id, message, client)
      user = @connected_clients[client_id]
      log_transaction(client_id, message)
    
      # Get the client's IP address
      client_ip = client.peeraddr[3]
    
      # Print the message with the client's IP address
      # puts "[#{Time.now}] (#{client_id}) <#{client_ip}> #{message}"
      puts "[#{Time.now}] <#{client_ip}> #{message}"

      if message.start_with?('@')
        process_command(client_id, message, client)
      elsif user.permissions.send
        broadcast_message(client_id, message)
      else
        send_error(client, "You don't have permission to send messages.")
      end
    end    

    def process_command(client_id, command, client)
      user = @connected_clients[client_id]
      puts "[#{Time.now}] <#{client_id}> (+c) #{command}"

      parts = command.split(' ')
      cmd = parts[0]

      case cmd
      when '@login_as_admin'
        login_as_admin(client_id, parts, client)
      when '@exit'
        disconnect_client(client_id)
      when '@broadcast'
        if user.permissions.broadcast
          broadcast_message(client_id, parts[1..-1].join(' '))
        else
          send_error(client, "You don't have permission to broadcast messages.")
        end
      when '@server-command'
        if user == @admin_user
          process_server_command(client_id, parts[1..-1].join(' '), client)
        else
          send_error(client, "Only admin can execute server commands.")
        end
      else
        send_error(client, "Unknown command.")
      end
    end

    def login_as_admin(client_id, parts, client)
      if parts.length != 3
        send_error(client, "Invalid login command. Use: @login_as_admin username password sessionId")
        return
      end

      username, password = parts[1], parts[2]

      if username == ServConf::Serv1::ADMIN_USER && password == ServConf::Serv1::PASSWORD
        @admin_user = @connected_clients[client_id]
        @admin_user.permissions = NetworkPermissions.new(read: true, write: true, execute: true, admin: true, send: true, receive: true, broadcast: true, configure: true, monitor: true, log: true, manage_connections: true, control_devices: true)
        client.puts "Login successful. You are now the admin."
      else
        send_error(client, "Invalid credentials or session ID.")
      end
    end

    def disconnect_client(client_id)
      if @connected_clients.key?(client_id)
        log_client_disconnection(client_id)
        @connected_clients.delete(client_id)
        puts "Client #{client_id} disconnected."
      end
    end

    def broadcast_message(sender_id, message)
      @connected_clients.each do |id, client|
        if client.permissions.receive && id != sender_id
          # In a real implementation, you'd have a way to write to each client's stream
          puts "Broadcast to #{client.username}: #{message}"
        end
      end
      log_transaction(sender_id, "BROADCAST: #{message}")
    end

    def process_server_command(client_id, command, client)
      parts = command.split(' ')
      cmd = parts[0]

      case cmd
      when 'userls'
        list_users(client)
      when 'user'
        if parts.length < 3
          send_error(client, "Invalid user command.")
          return
        end
        action, username = parts[1].downcase, parts[2]
        case action
        when 'ban'
          ban_user(username, "Banned by admin")
        when 'unban'
          unban_user(username)
        else
          send_error(client, "Invalid user action.")
        end
      when 'permissions'
        if parts.length < 2
          send_error(client, "Invalid permissions command.")
          return
        end
        show_user_permissions(parts[1], client)
      when 'update_permissions'
        if parts.length < 3
          send_error(client, "Invalid update_permissions command.")
          return
        end
        update_user_permissions(parts[1], parts[2..-1], client)
      when 'server'
        if parts.length < 2
          send_error(client, "Invalid server command.")
          return
        end
        case parts[1].downcase
        when 'stats'
          show_server_stats(client)
        when 'shutdown'
          shutdown_server
        else
          send_error(client, "Invalid server action.")
        end
      else
        send_error(client, "Unknown server command.")
      end
    end

    def list_users(client)
      client.puts "Connected users:"
      @connected_clients.each_value do |user|
        client.puts "- #{user.username}"
      end
    end

    def ban_user(username, reason)
      if @connected_clients.key?(username)
        @connected_clients.delete(username)
        File.open("./#{@current_session_id}/nodes/client/banned/list.json", 'a') do |f|
          f.puts({ Username: username, BannedAt: Time.now, Reason: reason }.to_json)
        end
        puts "User #{username} has been banned. Reason: #{reason}"
      end
    end

    def unban_user(username)
      banned_list_path = "./#{@current_session_id}/nodes/client/banned/list.json"
      banned_users = []
      if File.exist?(banned_list_path)
        banned_users = File.readlines(banned_list_path).map { |line| JSON.parse(line, symbolize_names: true) }
      end

      user_to_unban = banned_users.find { |u| u[:Username] == username }
      if user_to_unban
        banned_users.delete(user_to_unban)
        File.write(banned_list_path, banned_users.map(&:to_json).join("\n"))
        puts "User #{username} has been unbanned."
      else
        puts "User #{username} was not found in the banned list."
      end
    end

    def show_user_permissions(username, client)
      if @connected_clients.key?(username)
        user = @connected_clients[username]
        client.puts "Permissions for user #{username}:"
        user.permissions.print_permissions(client)
      else
        send_error(client, "User #{username} not found.")
      end
    end

    def update_user_permissions(username, permission_updates, client)
      if @connected_clients.key?(username)
        user = @connected_clients[username]
        permission_updates.each do |update|
          permission_name, value = update.split('=')
          if permission_name && value
            user.permissions.send("#{permission_name}=", value.downcase == 'true')
          end
        end
        client.puts "Updated permissions for user #{username}:"
        user.permissions.print_permissions(client)
      else
        send_error(client, "User #{username} not found.")
      end
    end

    def show_server_stats(client)
      client.puts "Server Statistics:"
      client.puts "Total connected users: #{@connected_clients.size}"
      client.puts "Total messages received: #{@message_counters.values.sum}"
      # Add more stats as needed
    end

    def shutdown_server
      puts "Server is shutting down..."
      @connected_clients.each_key { |client_id| disconnect_client(client_id) }
      @server.close
      exit(0)
    end

    def generate_session_id
      SecureRandom.uuid
    end

    def log_client_connection(client_id, ip_address)
      log_entry = {
        ClientId: client_id,
        IpAddress: ip_address,
        ConnectedAt: Time.now
      }.to_json

      File.open("./#{@current_session_id}/nodes/client/all/list.json", 'a') { |f| f.puts(log_entry) }
    end

    def log_client_disconnection(client_id)
      log_entry = {
        ClientId: client_id,
        DisconnectedAt: Time.now
      }.to_json

      File.open("./#{@current_session_id}/nodes/client/all/list.json", 'a') { |f| f.puts(log_entry) }
    end

    def log_transaction(client_id, message)
      transaction_id = generate_session_id
      transaction_dir = "./#{@current_session_id}/transactions/#{transaction_id}"

      # Ensure the directory exists
      FileUtils.mkdir_p(transaction_dir)

      log_entry = {
        ClientId: client_id,
        Message: message,
        Timestamp: Time.now
      }.to_json

      begin
        File.open("#{transaction_dir}/list.json", 'a') { |f| f.puts(log_entry) }
      rescue StandardError => e
        puts "Error logging transaction: #{e.message}"
      end
    end


    def send_error(client, error_message)
      client.puts "Error: #{error_message}"
      puts "Error sent to client: #{error_message}"
    end
  end

  class UserNode
    attr_reader :username
    attr_accessor :permissions

    def initialize(username, permissions)
      @username = username
      @permissions = permissions
    end
  end

  class NetworkPermissions
    attr_accessor :read, :write, :execute, :admin, :send, :receive, :broadcast, :configure, :monitor, :log, :manage_connections, :control_devices

    def initialize(read: false, write: false, execute: false, admin: false, send: false, receive: false, broadcast: false, configure: false, monitor: false, log: false, manage_connections: false, control_devices: false)
      @read = read
      @write = write
      @execute = execute
      @admin = admin
      @send = send
      @receive = receive
      @broadcast = broadcast
      @configure = configure
      @monitor = monitor
      @log = log
      @manage_connections = manage_connections
      @control_devices = control_devices
    end

    def print_permissions(client)
      instance_variables.each do |var|
        client.puts "#{var.to_s.delete('@')}: #{instance_variable_get(var)}"
      end
    end
  end

  module ServConf
    module Serv1
      ADMIN_USER = 'admin'
      PASSWORD = 'password123'
    end
  end
end

# Main program
server = MrRobotNet::NetworkServer.new(8080)
server.start