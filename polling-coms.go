package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

type IModbusIO interface {
	ReadInputs(address []uint16) ([]bool, error)
	ReadCoils(address []uint16) ([]bool, error)
	ReadAnalog(address []uint16) ([]float32, error)
	WriteCoil(address uint16, value bool) error
	Close() error
}

type IDataIO interface {
	OpenSocket(ip, port string) error
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

func pollLoop(ctx context.Context, plcConn IModbusIO, wConn IDataIO, addrRead *AddrMap, addrWrite *AddrMap, addrAnalog *AddrMap, pollPeriod time.Duration, uploadPeriod time.Duration) {
	ticker := time.NewTicker(pollPeriod)
	defer ticker.Stop()

	wFails := comFailures(InitWailonFails)
	plcFails := comFailures(InitModbusFails)

	uploadedAt := time.Now()
	readMemory := newReading(len(addrRead.logo), len(addrAnalog.logo))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
				continue
			}
			coilVals, err := plcConn.ReadCoils(coilAddrs)
			if err != nil {
				log.Printf("Error reading coils: %v", err)
				comFail(&plcFails)
				continue
			}
			anagVals, err := plcConn.ReadAnalog(addrAnalog.addr)
			if err != nil {
				log.Printf("Error reading analog inputs: %v", err)
				log.Print(addrAnalog.addr)
				comFail(&plcFails)
				continue
			}
			plcFails = InitModbusFails
			if !readMemory.ChangeInReading(append(inputVals, coilVals...)) &&
				!readMemory.ChangeInFloat(anagVals) &&
				!uploadedAt.Add(uploadPeriod).Before(time.Now()) {
				//log.Print("No change in registers")
				if err := wConn.SendPing(); err != nil {
					log.Printf("Error sending ping: %v", err)
					comFail(&wFails)
					continue
				}
				wFails = InitWailonFails
				continue
			}
			uploadedAt = time.Now()

			regReadings := append(inputVals, coilVals...)
			dataStr := ""
			for i, v := range regReadings {
				name := addrRead.name[i]
				value := 0
				if v {
					value = 1
				}
				if dataStr == "" {
					dataStr = fmt.Sprintf("%s:3:%d", name, value)
				} else {
					dataStr = fmt.Sprintf("%s:3:%d,%s", name, value, dataStr)
				}
			}
			bigWord := uint32(0)
			for j, val := range anagVals {
				name := addrAnalog.name[j]
				if addrAnalog.logo[j] == "0" {
					v := bigWord<<16 | uint32(val)
					bigWord = uint32(0)
					dataStr = fmt.Sprintf("%s:3:%.2f,%s", name, float64(v), dataStr)
				} else {
					bigWord = uint32(val)
				}
			}
			err = wConn.SendData(dataStr)
			if err != nil {
				log.Printf("Error: %v", err)
			}
		default:
			cmd, message, err := wConn.ReadCommand()
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}

			switch strings.ToUpper(cmd) {
			case "TIMEOUT":
				continue
			case "W":
				parts := strings.Split(message, "=")
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
						if err := plcConn.WriteCoil(addrWrite.addr[i], set); err != nil {
							log.Printf("error at %s=%t: %s", varName, set, err)
							continue
						}
						notFound = false
					}
				}
				if notFound {
					log.Printf("not found variable: %s", varName)
				}
			}
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

func (r *Reading) ChangeInReading(new []bool) bool {
	defer func() {
		copy(r.lastValues, new)
		r.sent = true
	}()
	if !r.sent {
		return true
	}
	for i := 0; i < len(new); i++ {
		if new[i] != r.lastValues[i] {
			return true
		}
	}
	return false
}

func (r *Reading) ChangeInFloat(new []float32) bool {
	defer func() {
		copy(r.lastFloats, new)
		r.sent = true
	}()
	if !r.sent {
		return true
	}

	for i, f := range new {
		if r.lastFloats[i] == 0 && f == 0 {
			continue
		}
		if r.lastFloats[i] == 0 && f != 0 {
			return true
		}
		diff := (f - r.lastFloats[i]) / r.lastFloats[i]
		if math.Abs(float64(diff)) > 0.1 {
			return true
		}
	}
	return false
}
