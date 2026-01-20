package wailonServer

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type WailonConnection struct {
	Imei string
	conn net.Conn
	mu   sync.Mutex
	Url  string
	Port string
}

func (c *WailonConnection) OpenSocket() error {

	if c.conn != nil {
		c.conn.Close()
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", c.Url, c.Port))
	c.conn = conn
	if err != nil {
		return fmt.Errorf("opening socket, got: %w", err)
	}

	login := fmt.Sprintf("2.0;%s;NA;", c.Imei)
	CRC := crcChecksum([]byte(login))
	res, err := writePacket(fmt.Sprintf("#L#%s%s\r\n", login, CRC), c.conn)
	if !strings.Contains(res, "#AL#1") {
		return fmt.Errorf("login unsuccessful, got: %s", res)
	}
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

	c.OpenSocket()

	t := time.Now()
	date := t.In(time.UTC).Format("020106")
	second := t.In(time.UTC).Format("150405")

	message := fmt.Sprintf("%s;%s;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;;NA;%s;", date, second, params)
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

	headerMessage := strings.Split(data, "#M#")
	if len(headerMessage) != 2 || headerMessage[0] != "" {
		return "", "", fmt.Errorf("should contain #M#: %s", data)
	}

	parts := strings.Split(headerMessage[1], "#")
	if len(parts) != 4 || parts[0] != "" {
		return "", "", fmt.Errorf("incomplete response: %s", headerMessage[1])
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
	if res, err := writePacket(fmt.Sprintf("#L#%s%s\r\n", login, CRC), c.conn); err != nil {
		return fmt.Errorf("writing to wailon, res: %s, got: %w", res, err)
	}
	//log.Printf("Ping, got %s", res)
	return nil
}
