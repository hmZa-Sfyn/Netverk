# Run: ab -n 10000 -c 100 http://localhost:3000/

####################################################################################### 
##   Author: hmZa-Sfyn                                                               ## 
##   Language: roobi-lang  (search that repo on my github `https://github.com/hmZa-Sfyn`)    ##
##      Version: 2024.69.1.2                                                         ##
##   Desc: A simple anti-ddos script for a web-server, written in roobi, not ruby!   ##
####################################################################################### 

require "net/simple_server"

server = Net::SimpleServer.new("6970")

i = 0

server.get("/") do |req, res|
  #puts(i)
  i = i+1

  if i % 20 == 0
    puts("[*] Starting Anto-DDOS Sleep.")
    sleep(1)
    puts("[!] Anto-DDOS Sleep Ended Gracefully.")
  end


  res.body = req.method + " --> hit"
  res.status = 200
end

server.start