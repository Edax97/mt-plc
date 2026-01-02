package main

import (
	"flag"
	"log"
	"mt-plc-control/modbusClient"
	"time"
)

const AddrModbus = "192.168.8.52"
const PortModbus = "502"

func main() {
	plcConn, err := modbusClient.NewModbusConn(AddrModbus+":"+PortModbus, 2000*time.Millisecond)

	if err != nil {
		panic(err)
	}

	cmd := flag.String("cmd", "OFF", "Comando a enviar")
	flag.Parse()
	if cmd == nil {
		panic("Nil flag cmd")
	}

	if *cmd == "ON" {
		if err := modbusClient.GenSetON(plcConn); err != nil {
			panic(err)
		}
	} else if *cmd == "OFF" {
		if err := modbusClient.GenSetOFF(plcConn); err != nil {
			panic(err)
		}
	}

	log.Printf("Successfully executed cmd %s on GENSET", *cmd)
}
