package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mt-plc-control/modbusClient"
	"mt-plc-control/wailonServer"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AddrMap struct {
	logo []string
	name []string
	addr []uint16
}

func ParseAddrMap(fields string) *AddrMap {
	lines := strings.Split(fields, "\n")
	logo := make([]string, 0)
	name := make([]string, 0)
	addr := make([]uint16, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			return nil
		}
		logo = append(logo, parts[0])
		name = append(name, parts[1])

		aInt, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil
		}
		addr = append(addr, uint16(aInt))
	}
	return &AddrMap{logo, name, addr}
}

// getBit extrae un bit (posición bitIndex) del slice de bytes devuelto por Modbus

func main() {
	//ENV
	_ = godotenv.Load()
	Imei := os.Getenv("IMEI")
	isMock := os.Getenv("MOCK")
	AddrModbus := os.Getenv("ADDR_MODBUS")
	PortModbus := os.Getenv("PORT_MODBUS")
	UrlWailon := os.Getenv("URL_WAILON")
	PortWailon := os.Getenv("PORT_WAILON")
	addrRead := ParseAddrMap(os.Getenv("REGISTERS_READ"))
	addrWrite := ParseAddrMap(os.Getenv("REGISTERS_WRITE"))
	addrAnalog := ParseAddrMap(os.Getenv("REGISTERS_ANALOG"))

	if addrRead == nil || addrWrite == nil {
		log.Fatal("Malformed REGISTER_READ(WRITE)")
	}

	timeoutMs, err := strconv.Atoi(os.Getenv("TIMEOUT_MODBUS"))
	if err != nil {
		timeoutMs = 2500
	}
	period := flag.Int("period", 10, "Periodo de polling en s")
	uploadMin := flag.Int("uploadMin", 10, "Periodo para upload en min")
	flag.Parse()

	plcAddr := fmt.Sprintf("%s:%s", AddrModbus, PortModbus)
	plcConn, err := modbusClient.NewModbusConn(plcAddr, time.Duration(timeoutMs)*time.Millisecond)
	if err != nil {
		log.Fatalf("no se pudo conectar al PLC en %s: %v", plcAddr, err)
	}
	defer func() {
		_ = plcConn.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wailonCon IDataIO

	if isMock == "1" {
		wailonCon = wailonServer.NewMockServer()
	} else {
		wailonCon = &wailonServer.WailonConnection{Imei: Imei}
	}
	err = wailonCon.OpenSocket(UrlWailon, PortWailon)
	if err != nil {
		log.Fatalf("no se pudo conectar al servidor wailon: %v", err)
	}
	defer func() {
		wailonCon.CloseSocket()
	}()

	log.Printf("Conectado a %s", UrlWailon)
	pollLoop(ctx, plcConn, wailonCon, addrRead, addrWrite, addrAnalog,
		time.Duration(*period)*time.Second,
		time.Duration(*uploadMin)*time.Minute)
}
