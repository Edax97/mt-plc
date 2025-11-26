package main

import (
	"time"

	"github.com/goburrow/modbus"
)

type modbusConn struct {
	handler *modbus.TCPClientHandler
	client  modbus.Client
}

func NewModbusConn(address string, timeout time.Duration) (IModbusIO, error) {
	h := modbus.NewTCPClientHandler(address)
	h.Timeout = timeout
	h.SlaveId = 1 // LOGO! por defecto usa ID 1 cuando está detrás de TCP gateway
	if err := h.Connect(); err != nil {
		return nil, err
	}
	return &modbusConn{handler: h, client: modbus.NewClient(h)}, nil
}

// Write unit tests for methods in this file
func (c *modbusConn) ReadInputs(addressList []uint16) ([]bool, error) {
	inputBool := make([]bool, len(addressList))
	iStart := extremeValue(addressList, min16)
	iEnd := extremeValue(addressList, max16)
	iQty := iEnd - iStart + 1
	iRegs, err := c.client.ReadDiscreteInputs(iStart, iQty)
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

	qStart := extremeValue(addressList, min16)
	qEnd := extremeValue(addressList, max16)
	qQty := qEnd - qStart + 1

	qRegs, err := c.client.ReadCoils(qStart, qQty)
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
	aStart := extremeValue(addressList, min16)
	aEnd := extremeValue(addressList, max16)
	aQty := aEnd - aStart + 1

	bytesArr, err := c.client.ReadInputRegisters(aStart-1, aQty)

	if err != nil {
		return nil, err
	}
	analogs := make([]float32, len(addressList))
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
	_, err := c.client.WriteSingleCoil(address, v)
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
