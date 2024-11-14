require "net/http"

c = Channel.new

10001.times do    
    puts(Net::HTTP.get("http://127.0.0.1:6970"))
end

c.close
