package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// Ping statistics
type PingStats struct {
	Sent      int
	Received  int
	Lost      int
	MinRTT    time.Duration
	MaxRTT    time.Duration
	TotalRTT  time.Duration
	RTTValues []time.Duration
}

// Default packet size for ping
const DefaultPacketSize = 32

// Resolve hostname to IP address
func resolveHostname(hostname string) (net.IP, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}
	return ips[0], nil
}

// Ping a host using ICMP Echo Request
func ping(ip net.IP, packetSize int, timeout time.Duration, ttl int, protocol string) (time.Duration, error) {
	var conn *icmp.PacketConn
	var err error

	var icmpTypeEcho icmp.Type
	var icmpTypeEchoReply icmp.Type

	if protocol == "ipv4" {
		conn, err = icmp.ListenPacket("ip4:icmp", "")
		icmpTypeEcho = ipv4.ICMPTypeEcho
		icmpTypeEchoReply = ipv4.ICMPTypeEchoReply
	} else if protocol == "ipv6" {
		conn, err = icmp.ListenPacket("ip6:ipv6-icmp", "")
		icmpTypeEcho = ipv6.ICMPTypeEchoRequest
		icmpTypeEchoReply = ipv6.ICMPTypeEchoReply
	} else {
		return 0, fmt.Errorf("unsupported protocol: %s", protocol)
	}
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	if protocol == "ipv4" {
		conn.IPv4PacketConn().SetTTL(ttl)
	}

	// Create ICMP Echo Request message
	echo := icmp.Message{
		Type: icmpTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: make([]byte, packetSize),
		},
	}

	// Marshal the message into binary
	msgBytes, err := echo.Marshal(nil)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: ip})
	if err != nil {
		return 0, err
	}

	// Set read timeout
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return 0, err
	}

	// Buffer to receive the reply
	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return 0, err
	}

	duration := time.Since(start)
	// Parse the reply
	rm, err := icmp.ParseMessage(icmpTypeEchoReply.Protocol(), reply[:n])
	if err != nil {
		return 0, err
	}

	switch rm.Type {
	case icmpTypeEchoReply:
		return duration, nil
	default:
		return 0, fmt.Errorf("got unexpected ICMP type: %v", rm.Type)
	}
}

// Calculate ping statistics
func calculateStats(stats *PingStats) {
	if len(stats.RTTValues) == 0 {
		return
	}

	min := stats.RTTValues[0]
	max := stats.RTTValues[0]
	total := time.Duration(0)
	for _, rtt := range stats.RTTValues {
		if rtt < min {
			min = rtt
		}
		if rtt > max {
			max = rtt
		}
		total += rtt
	}

	stats.MinRTT = min
	stats.MaxRTT = max
	stats.TotalRTT = total

	mean := total / time.Duration(len(stats.RTTValues))

	variance := 0.0
	for _, rtt := range stats.RTTValues {
		diff := float64(rtt - mean)
		variance += diff * diff
	}
	stdDev := time.Duration(math.Sqrt(variance / float64(len(stats.RTTValues))))
	fmt.Printf("Statistics: \nMin RTT: %v \nMax RTT: %v \nAverage RTT: %v \nStandard Deviation: %v \n", stats.MinRTT, stats.MaxRTT, mean, stdDev)
}

func main() {
	// Command-line flags
	host := flag.String("host", "", "Host to ping (IP or hostname)")
	packetSize := flag.Int("s", DefaultPacketSize, "Size of packet in bytes")
	count := flag.Int("c", 4, "Number of pings to send")
	interval := flag.Duration("i", 1*time.Second, "Interval between pings")
	timeout := flag.Duration("t", 2*time.Second, "Timeout for each ping")
	ttl := flag.Int("ttl", 64, "TTL (Time To Live) value")
	protocol := flag.String("proto", "ipv4", "Protocol (ipv4 or ipv6)")
	flag.Parse()

	if *host == "" {
		log.Fatal("Please provide a host to ping using -host <hostname or IP>")
	}

	ip, err := resolveHostname(*host)
	if err != nil {
		log.Fatalf("Failed to resolve hostname: %v", err)
	}

	fmt.Printf("PING %s (%s) with %d bytes of data:\n", *host, ip, *packetSize)

	stats := PingStats{}

	for i := 0; i < *count; i++ {
		stats.Sent++
		rtt, err := ping(ip, *packetSize, *timeout, *ttl, *protocol)
		if err != nil {
			stats.Lost++
			fmt.Printf("Request timeout for icmp_seq %d\n", i+1)
		} else {
			stats.Received++
			stats.RTTValues = append(stats.RTTValues, rtt)
			fmt.Printf("%d: %d bytes from %s: icmp_seq=%d time=%v TTL=%d\n", i, *packetSize, ip, i+1, rtt, *ttl)
		}

		time.Sleep(*interval)
	}

	// Print summary statistics
	fmt.Printf("\n--- %s ping statistics ---\n", *host)
	fmt.Printf("%d packets transmitted, %d packets received, %.1f%% packet loss\n", stats.Sent, stats.Received, float64(stats.Lost)/float64(stats.Sent)*100)
	calculateStats(&stats)
}
