package wailonServer

import (
	"fmt"
	"log"
	"time"
)

type MockServer struct {
}

func NewMockServer() *MockServer {
	return &MockServer{}
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
