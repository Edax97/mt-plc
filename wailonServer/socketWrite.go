package wailonServer

import (
	"bufio"
	"net"
	"time"
)

func writePacket(packet string, con net.Conn) (string, error) {
	_, err := con.Write([]byte(packet))
	if err != nil {
		return "", err
	}

	_ = con.SetReadDeadline(time.Now().Add(2000 * time.Millisecond))
	defer func() {
		_ = con.SetReadDeadline(time.Time{})
	}()
	reader := bufio.NewReader(con)
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return res, nil
}

func readPacket(con net.Conn) (string, error) {
	_ = con.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	defer func() {
		_ = con.SetReadDeadline(time.Time{})
	}()
	reader := bufio.NewReader(con)
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return res, nil
}
