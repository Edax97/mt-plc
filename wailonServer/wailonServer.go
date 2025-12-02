package wailonServer

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type WailonConnection struct {
	Imei string
	conn net.Conn
	mu   sync.Mutex
}

func (c *WailonConnection) OpenSocket(ip, port string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ip, port))
	c.conn = conn
	if err != nil {
		return fmt.Errorf("opening socket, got: %w", err)
	}

	login := fmt.Sprintf("2.0;%s;NA;", c.Imei)
	CRC := crcChecksum([]byte(login))
	_, err = writePacket(fmt.Sprintf("#L#%s%s\r\n", login, CRC), c.conn)
	if err != nil {
		return fmt.Errorf("on login, got: %w", err)
	}

	return nil
}

func (c *WailonConnection) CloseSocket() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// param:type:value,param2
func (c *WailonConnection) SendData(params string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	message := fmt.Sprintf("NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;;NA;%s", params)
	CRC := crcChecksum([]byte(message))
	res, err := writePacket(fmt.Sprintf("#D#%s%s\r\n", message, CRC), c.conn)
	if err != nil {
		return fmt.Errorf("when writing to wailon, got: %w \nsent:%s", err, message)
	}

	if !strings.Contains(res, "#AD#1") {
		return fmt.Errorf("response unsuccessful, got: %s", res)
	}
	log.Printf("Sent %s, got %s", message, res)
	return nil
}

func (c *WailonConnection) ReadCommand() (kind string, value string, e error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := readPacket(c.conn)
	if data == "" {
		return "Timeout", "", nil
	}
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(data, "#")
	if len(parts) != 4 {
		return "", "", fmt.Errorf("incomplete response: %s", data)
	}
	kind = parts[1]
	value = parts[2]
	return
}

func (c *WailonConnection) SendPing() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	login := fmt.Sprintf("2.0;%s;NA;", c.Imei)
	CRC := crcChecksum([]byte(login))
	res, err := writePacket(fmt.Sprintf("#L#%s%s\r\n", login, CRC), c.conn)
	if err != nil {
		return fmt.Errorf("when writing to wailon, got: %w", err)
	}
	log.Printf("Ping, got %s", res)
	return nil
}
