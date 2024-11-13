package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

var portDescriptions = map[int]string{
	// Well-known ports (0–1023)
	20:   "FTP Data Transfer",
	21:   "FTP Control",
	22:   "SSH Remote Login",
	23:   "Telnet",
	25:   "SMTP Email Routing",
	53:   "DNS (Domain Name System)",
	67:   "DHCP (Server)",
	68:   "DHCP (Client)",
	69:   "TFTP (Trivial File Transfer Protocol)",
	80:   "HTTP (Hypertext Transfer Protocol)",
	110:  "POP3 (Post Office Protocol)",
	119:  "NNTP (Network News Transfer Protocol)",
	123:  "NTP (Network Time Protocol)",
	135:  "RPC (Remote Procedure Call)",
	143:  "IMAP (Internet Message Access Protocol)",
	161:  "SNMP (Simple Network Management Protocol)",
	194:  "IRC (Internet Relay Chat)",
	389:  "LDAP (Lightweight Directory Access Protocol)",
	443:  "HTTPS (HTTP Secure)",
	445:  "Microsoft-DS (Active Directory, SMB)",
	465:  "SMTPS (Simple Mail Transfer Protocol Secure)",
	500:  "ISAKMP (VPN)",
	514:  "Syslog",
	520:  "RIP (Routing Information Protocol)",
	546:  "DHCPv6 Client",
	547:  "DHCPv6 Server",
	587:  "SMTP (Mail Submission)",
	631:  "IPP (Internet Printing Protocol)",
	993:  "IMAPS (Secure IMAP)",
	995:  "POP3S (Secure POP3)",
	1433: "Microsoft SQL Server",
	1434: "Microsoft SQL Monitor",
	1521: "Oracle Database",
	1701: "L2TP (Layer 2 Tunneling Protocol)",
	1723: "PPTP (Point-to-Point Tunneling Protocol)",
	1812: "RADIUS Authentication",
	1813: "RADIUS Accounting",
	2049: "NFS (Network File System)",
	2082: "cPanel (Web Hosting Management)",
	2083: "cPanel (Web Hosting Secure)",
	2100: "Oracle XDB FTP",
	2222: "DirectAdmin (Control Panel)",
	2483: "Oracle DB listener",
	2484: "Oracle DB listener (Secure)",
	3306: "MySQL Database",
	3389: "Microsoft RDP (Remote Desktop Protocol)",
	3690: "Subversion (SVN)",
	4444: "Metasploit RPC Server",
	4567: "CruiseControl",
	4662: "eMule (P2P file sharing)",
	5432: "PostgreSQL Database",
	5900: "VNC (Virtual Network Computing)",
	6379: "Redis",
	6660: "IRC (Internet Relay Chat)",
	6881: "BitTorrent (P2P File Sharing)",
	8080: "HTTP Proxy/Alternative HTTP",
	8443: "HTTPS Alternative",
	8888: "HTTP Alternative",
	9000: "SonarQube",
	9090: "Openfire Administration Console",
	9200: "Elasticsearch",
	11211: "Memcached",
	27017: "MongoDB",
	27018: "MongoDB (Replica Set) (Secondary)",
	27019: "MongoDB (Shard) (Arbiter)",
	50000: "SAP Management Console",
	50030: "Hadoop JobTracker",
	50070: "Hadoop NameNode",
	54321: "BOINC (Distributed Computing)",
	55000: "Veeam Backup",
	636:  "LDAPS (Secure LDAP)",
	873:  "Rsync",
	902:  "VMware Server Console",
	1080: "SOCKS Proxy",
	1194: "OpenVPN",
	3260: "iSCSI (Internet Small Computer System Interface)",
}

func main() {
	app := &cli.App{
		Name:  "netshell",
		Usage: "Enhanced network scanning tool with CIDR support",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "on",
				Usage:    "CIDR notation of the network to scan (e.g., 192.168.0.0/24)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "listusers",
				Usage: "List connected devices",
			},
			&cli.BoolFlag{
				Name:  "scanports",
				Usage: "Scan for open ports on all connected devices",
			},
			&cli.IntFlag{
				Name:  "portrange",
				Usage: "Set the range of ports to scan (use with --scanports)",
				Value: 1024,
			},
		},
		Action: func(c *cli.Context) error {
			cidr := c.String("on")
			devices := []string{}

			if c.Bool("listusers") {
				fmt.Printf("Scanning %s for devices!\n", cidr)
				devices = listConnectedDevices(cidr)
			}

			if c.Bool("scanports") {
				if len(devices) == 0 {
					devices = listConnectedDevices(cidr)
				}
				for _, device := range devices {
					scanOpenPorts(device, c.Int("portrange"))
				}
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func listConnectedDevices(cidr string) []string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println("Invalid CIDR notation:", err)
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	devices := []string{}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			addr, err := net.LookupAddr(ip)
			if err == nil {
				mu.Lock()
				devices = append(devices, ip)
				mu.Unlock()
				if len(addr) > 0 {
					fmt.Printf("* %s (%s)\n", ip, addr[0])
				} else {
					fmt.Printf("* %s\n", ip)
				}
			}
		}(ip.String())
	}
	wg.Wait()
	return devices
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func scanOpenPorts(ip string, portRange int) {
	var wg sync.WaitGroup
	for port := 1; port <= portRange; port++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
			if err == nil {
				conn.Close()
				portUse := "Unknown"
				if description, ok := portDescriptions[port]; ok {
					portUse = description
				}
				fmt.Printf("Port %d open on %s (%s)\n", port, ip, portUse)
			}
		}(port)
	}
	wg.Wait()
}
