package main

import (
	"flag"
	"fmt"
	"log"
	"mt-plc-control/modbusClient"
	"time"
)

const AddrModbus = "192.168.8.52"
const PortModbus = "502"

func main() {
	plcConn, err := modbusClient.NewModbusConn(AddrModbus+":"+PortModbus, 4000*time.Millisecond)

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
	} else if *cmd == "EN" {

		if err := plcConn.WriteCoil(4700, true); err != nil {
			val, _ := plcConn.ReadCoils([]uint16{4700})
			fmt.Printf("values from 4700: %v \n", val)
			panic(err)
		}
	} else if *cmd == "DI" {

		if err := plcConn.WriteCoil(4700, false); err != nil {
			val, _ := plcConn.ReadCoils([]uint16{4700})
			fmt.Printf("values from 4700: %v \n", val)
			panic(err)
		}
	}

	log.Printf("Successfully executed cmd %s on GENSET", *cmd)
}
