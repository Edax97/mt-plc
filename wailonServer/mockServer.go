package wailonServer

import (
	"fmt"
	"log"
	"time"
)

type MockServer struct {
	ip   string
	port string
}

func (s *MockServer) OpenSocket() error {
	log.Printf("serverd mocked at %s:%s", s.ip, s.port)
	return nil
}

func (s *MockServer) SendPing() error {
	log.Println("ping...")
	return nil
}

func (s *MockServer) CloseSocket() {
}

func (s *MockServer) SendData(params string) error {
	t := time.Now()
	date := t.In(time.UTC).Format("020106")
	second := t.In(time.UTC).Format("150405")

	message := fmt.Sprintf("%s;%s;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;;NA;%s;", date, second, params)
	CRC := crcChecksum([]byte(message))
	packet := fmt.Sprintf("#D#%s%s\r\n", message, CRC)
	fmt.Printf("Sending... %s", packet)
	return nil
}

func (s *MockServer) ReadCommand() (string, string, error) {
	time.Sleep(time.Second * 5)

	fmt.Println("Hi")

	t := time.Now()

	if t.Second()%7 > 5 {
		return "W", "W_Q1=1", nil
	} else {
		return "Timeout", "", nil
	}

}

func NewMockServer(ip, port string) *MockServer {
	return &MockServer{ip, port}
}

func (s *MockServer) SendTimeValue(imei string, date time.Time, wh string, vah string, vao string) (bool, error) {
	login := fmt.Sprintf("2.0;%s;NA;", imei)
	CRC := crcChecksum([]byte(login))
	loginPacket := fmt.Sprintf("#L#%s%s\r\n", login, CRC)
	log.Printf("  - IMEI: %s\n    LOGIN PACKET: %s\n", imei, loginPacket)

	hourStr := date.In(time.UTC).Format("2006.01.02.15.04")
	data := fmt.Sprintf("WH:3:%s,VARH:3:%s,VAO:3:%s;", wh, vah, vao)
	message := fmt.Sprintf("%s;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;NA;;NA;%s", hourStr, data)
	CRC = crcChecksum([]byte(message))
	dataPacket := fmt.Sprintf("#D#%s%s\r\n", message, CRC)
	fmt.Printf("  - DATA PACKET: %s\n", dataPacket)

	return true, nil
}
