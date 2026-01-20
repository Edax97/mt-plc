package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"mt-plc-control/modbusClient"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type IDataIO interface {
	OpenSocket() error
	SendPing() error
	CloseSocket()
	SendData(params string) error
	ReadCommand() (string, string, error)
}

const (
	InitModbusFails = 6
	InitWailonFails = 5
)

type comFailures int

func comFail(f *comFailures) {
	if f == nil {
		return
	}
	*f = *f - 1
	if *f <= 0 {
		os.Exit(1)
	}
}

func pollLoop(ctx context.Context, plcConn *modbusClient.ModbusConn, wConn IDataIO, addrRead *AddrMap, addrWrite *AddrMap, addrAnalog *AddrMap, pollPeriod time.Duration, uploadPeriod time.Duration) {
	_ = godotenv.Load()
	ticker := time.NewTicker(pollPeriod)
	defer ticker.Stop()

	wFails := comFailures(InitWailonFails)
	plcFails := comFailures(InitModbusFails)

	uploadedAt := time.Now()
	readMemory := newReading(len(addrRead.logo), len(addrAnalog.logo))

	sendData := func(sendNow bool) {
		//log.Print("Tick")
		inputAddrs := make([]uint16, 0)
		coilAddrs := make([]uint16, 0)

		for j, v := range addrRead.logo {
			w := []rune(v)
			switch w[0] {
			case 'I':
				inputAddrs = append(inputAddrs, addrRead.addr[j])
			case 'Q':
				coilAddrs = append(coilAddrs, addrRead.addr[j])
			}
		}

		inputVals, err := plcConn.ReadInputs(inputAddrs)
		if err != nil {
			log.Printf("Error reading inputs: %v", err)
			comFail(&plcFails)
			return
		}
		time.Sleep(time.Millisecond * 50)
		coilVals, err := plcConn.ReadCoils(coilAddrs)
		if err != nil {
			log.Printf("Error reading coils: %v", err)
			comFail(&plcFails)
			return
		}
		time.Sleep(time.Millisecond * 50)
		anagVals, err := plcConn.ReadAnalog(addrAnalog.addr)
		if err != nil {
			log.Printf("Error reading analog inputs: %v", err)
			log.Print(addrAnalog.addr)
			comFail(&plcFails)
			return
		}
		time.Sleep(time.Millisecond * 50)
		plcFails = InitModbusFails
		if !sendNow && !readMemory.HaveChanged(coilVals, anagVals) &&
			!uploadedAt.Add(uploadPeriod).Before(time.Now()) {
			//log.Print("No change in registers")
			if err := wConn.SendPing(); err != nil {
				log.Printf("Error sending ping: %v", err)
				comFail(&wFails)
				return
			}
			wFails = InitWailonFails
			return
		}
		uploadedAt = time.Now()
		readMemory.UpdateLastValues(coilVals, anagVals)

		regReadings := append(inputVals, coilVals...)
		dataStr := ""
		for i, v := range regReadings {
			name := addrRead.name[i]
			value := 0
			if v {
				value = 1
			}
			if dataStr == "" {
				dataStr = fmt.Sprintf("%s:1:%d", name, value)
			} else {
				dataStr = fmt.Sprintf("%s:1:%d,%s", name, value, dataStr)
			}
		}
		bigWord := uint32(0)
		for j, val := range anagVals {
			name := addrAnalog.name[j]
			if addrAnalog.logo[j] == "0" {
				v := bigWord<<16 | uint32(val)
				bigWord = uint32(0)
				dataStr = fmt.Sprintf("%s:1:%d,%s", name, v, dataStr)
			} else {
				bigWord = uint32(val)
			}
		}
		err = wConn.SendData(dataStr)
		if err != nil {
			log.Printf("Error: %v", err)
		}

	}

	cmdChan := make(chan string)
	go func() {
		for {
			cmd, message, err := wConn.ReadCommand()
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			if strings.ToUpper(cmd) == "TIMEOUT" {
				continue
			}
			cmdChan <- fmt.Sprintf("%s|%s", cmd, message)
		}

	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendData(false)
		case cmd := <-cmdChan:
			cmdParts := strings.Split(cmd, "|")
			code, message := cmdParts[0], cmdParts[1]
			switch strings.ToUpper(code) {
			case "TIMEOUT":
				continue
			case "W":
				for line := range strings.SplitSeq(message, ";") {
					parts := strings.Split(line, "=")
					if len(parts) != 2 {
						log.Printf("malformed command: %s", cmd)
						continue
					}
					varName := strings.Trim(parts[0], " ")
					set := false
					if strings.Trim(parts[1], " \r\n") == "1" {
						set = true
					}
					log.Printf("Comand %s=%t", varName, set)

					notFound := true
					for i, regName := range addrWrite.name {
						if varName == regName {
							notFound = false
							if err := plcConn.WriteCoil(addrWrite.addr[i], set); err != nil {
								log.Printf("error at %s=%t: %s", varName, set, err)
							}
							time.Sleep(time.Millisecond * 50)
						}
					}
					if notFound {
						log.Printf("not found variable: %s", varName)
					}

				}
			case "GS":
				en := os.Getenv("GENSET_COMMANDS")
				if en != "true" {
					continue
				}
				if message == "START" {
					if err := modbusClient.GenSetON(plcConn); err != nil {
						log.Printf("Error prendiendo gen %v", err)
						continue
					}
				} else if message == "STOP" {
					if err := modbusClient.GenSetOFF(plcConn); err != nil {
						log.Printf("Error apagando gen %v", err)
						continue
					}
				}
			}
			time.Sleep(time.Millisecond * 500)
			sendData(true)
		}
	}
}

type Reading struct {
	lastValues []bool
	lastFloats []float32
	sent       bool
}

func newReading(s, sf int) *Reading {
	l := make([]bool, s)
	f := make([]float32, sf)
	return &Reading{l, f, false}
}

func (r *Reading) HaveChanged(b []bool, f []float32) bool {

	if !r.sent {
		return true
	}

	for i := 0; i < len(b); i++ {
		if b[i] != r.lastValues[i] {
			return true
		}
	}
	for j := 0; j < len(f); j++ {
		if r.lastFloats[j] == 0 {
			if f[j] != 0 {
				return true
			}
		} else {
			diff := (f[j] - r.lastFloats[j]) / r.lastFloats[j]
			if math.Abs(float64(diff)) > 0.1 {
				return true
			}
		}
	}
	return false
}

func (r *Reading) UpdateLastValues(b []bool, f []float32) {
	copy(r.lastFloats, f)
	copy(r.lastValues, b)
	r.sent = true
}
