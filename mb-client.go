package main

import (
	"fmt"
	"time"

	"github.com/goburrow/modbus"
)

type modbusConn struct {
	handler *modbus.TCPClientHandler
}

func NewModbusConn(address string, timeout time.Duration) (IModbusIO, error) {
	h := modbus.NewTCPClientHandler(address)
	h.Timeout = timeout
	h.SlaveId = 1 // LOGO! por defecto usa ID 1 cuando está detrás de TCP gateway
	return &modbusConn{handler: h}, nil
}

func (c *modbusConn) StartConnection() (modbus.Client, error) {
	if err := c.handler.Connect(); err != nil {
		return nil, err
	}
	return modbus.NewClient(c.handler), nil
}

func (c *modbusConn) ReadInputs(addressList []uint16) ([]bool, error) {
	inputBool := make([]bool, len(addressList))
	if len(addressList) == 0 {
		return inputBool, nil
	}
	iStart := extremeValue(addressList, min16)
	iEnd := extremeValue(addressList, max16)
	iQty := iEnd - iStart + 1

	client, err := c.StartConnection()
	if err != nil {
		return []bool{}, fmt.Errorf("when connecting, %w", err)
	}
	defer func() {
		_ = c.Close()
	}()
	iRegs, err := client.ReadDiscreteInputs(iStart, iQty)
	if err != nil {
		return nil, err
	}

	for j, a := range addressList {
		b := getBit(iRegs, a-iStart)
		inputBool[j] = b
	}
	return inputBool, nil
}

func (c *modbusConn) ReadCoils(addressList []uint16) ([]bool, error) {
	coilsBool := make([]bool, len(addressList))
	if len(addressList) == 0 {
		return coilsBool, nil
	}

	qStart := extremeValue(addressList, min16)
	qEnd := extremeValue(addressList, max16)
	qQty := qEnd - qStart + 1

	client, err := c.StartConnection()
	if err != nil {
		return []bool{}, fmt.Errorf("when connecting, %w", err)
	}
	defer func() {
		_ = c.Close()
	}()
	qRegs, err := client.ReadCoils(qStart, qQty)
	if err != nil {
		return nil, err
	}

	for j, a := range addressList {
		b := getBit(qRegs, a-qStart)
		coilsBool[j] = b
	}
	return coilsBool, nil
}

func (c *modbusConn) ReadAnalog(addressList []uint16) ([]float32, error) {
	analogs := make([]float32, len(addressList))
	if len(addressList) == 0 {
		return analogs, nil
	}
	aStart := extremeValue(addressList, min16)
	aEnd := extremeValue(addressList, max16)
	aQty := aEnd - aStart + 1

	client, err := c.StartConnection()
	if err != nil {
		return []float32{}, fmt.Errorf("when connecting, %w", err)
	}
	defer func() {
		_ = c.Close()
	}()
	bytesArr, err := client.ReadInputRegisters(aStart, aQty)

	if err != nil {
		return nil, err
	}
	for j, add := range addressList {
		analogs[j] = getFloat(bytesArr, add-aStart)
	}
	return analogs, nil
}

func (c *modbusConn) WriteCoil(address uint16, value bool) error {
	var v uint16
	if value {
		v = 0xFF00
	} else {
		v = 0x0000
	}
	client, err := c.StartConnection()
	if err != nil {
		return fmt.Errorf("when connecting, %w", err)
	}
	defer func() {
		_ = c.Close()
	}()
	_, err = client.WriteSingleCoil(address, v)
	return err
}

func (c *modbusConn) Close() error {
	if c.handler != nil {
		return c.handler.Close()
	}
	return nil
}

func min16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
func max16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}
func extremeValue(l []uint16, comp func(i, j uint16) uint16) uint16 {
	extreme := l[0]
	for _, v := range l {
		extreme = comp(extreme, v)
	}
	return extreme
}

func getBit(data []byte, bitIndex uint16) bool {
	byteIndex := int(bitIndex / 8)
	bitPos := uint(bitIndex % 8)
	if byteIndex < 0 || byteIndex >= len(data) {
		return false
	}
	return (data[byteIndex] & (1 << bitPos)) != 0
}

func getFloat(data []byte, addrIndex uint16) float32 {
	byteIndex := int(addrIndex * 2)
	intScale := uint16(data[byteIndex])<<8 | uint16(data[byteIndex+1])
	//floatScale := float32(intScale-1<<8) * 0.1
	floatScale := float32(intScale)
	return floatScale
}
