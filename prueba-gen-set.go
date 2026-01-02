package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	AddrModbus := os.Getenv("ADDR_MODBUS")
	PortModbus := os.Getenv("PORT_MODBUS")

	plcConn, err := NewModbusConn(AddrModbus+":"+PortModbus, 2000*time.Millisecond)

	if err != nil {
		panic(err)
	}

	cmd := flag.String("cmd", "OFF", "Comando a enviar")
	flag.Parse()
	if cmd == nil {
		panic("Nil flag cmd")
	}

	if *cmd == "ON" {
		if err := GenSetON(plcConn); err != nil {
			panic(err)
		}
	} else if *cmd == "OFF" {
		if err := GenSetOFF(plcConn); err != nil {
			panic(err)
		}
	}

	log.Printf("Successfully executed cmd %s on GENSET", *cmd)
}
