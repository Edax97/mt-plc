package modbusClient

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/goburrow/modbus"
)

type ModbusConn struct {
	handler *modbus.TCPClientHandler
}

func NewModbusConn(address string, timeout time.Duration) (*ModbusConn, error) {
	h := modbus.NewTCPClientHandler(address)
	h.Timeout = timeout
	h.SlaveId = 1 // LOGO! por defecto usa ID 1 cuando está detrás de TCP gateway
	return &ModbusConn{handler: h}, nil
}

func (c *ModbusConn) StartConnection() (modbus.Client, error) {
	if err := c.handler.Connect(); err != nil {
		return nil, err
	}
	return modbus.NewClient(c.handler), nil
}

func (c *ModbusConn) ReadInputs(addressList []uint16) ([]bool, error) {
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

func (c *ModbusConn) ReadCoils(addressList []uint16) ([]bool, error) {
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

func (c *ModbusConn) ReadAnalog(addressList []uint16) ([]float32, error) {
	analogs := make([]float32, len(addressList))
	bytesArr := make([]byte, 0, 2*len(addressList))

	if len(addressList) == 0 {
		return analogs, nil
	}

	client, err := c.StartConnection()
	if err != nil {
		return []float32{}, fmt.Errorf("when connecting, %w", err)
	}
	defer func() {
		_ = c.Close()
	}()

	aData := struct {
		aStart uint16
		aQty   uint16
	}{addressList[0], 1}
	for j := 1; j < len(addressList)+1; j++ {
		var aj uint16
		if j < len(addressList) {
			aj = addressList[j]
		}
		if aj > aData.aStart+aData.aQty || j == len(addressList) {
			b, err := client.ReadInputRegisters(aData.aStart, aData.aQty)
			if err != nil {
				return nil, err
			}
			bytesArr = append(bytesArr, b...)
			aData.aStart = aj
			aData.aQty = 1
		} else {
			aData.aQty = aData.aQty + 1
		}
	}

	for i := range addressList {
		analogs[i] = getFloat(bytesArr, uint16(i))
	}

	return analogs, nil
}

/*
func (c *ModbusConn) ReadAnalog(addressList []uint16) ([]float32, error) {
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
}*/

func (c *ModbusConn) WriteCoil(address uint16, value bool) error {
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

func (c *ModbusConn) WriteCommand(cmdAddress uint16, cmdValue uint16, argAddress uint16, argValue uint32) (uint32, error) {
	client, err := c.StartConnection()
	defer func() {
		_ = c.Close()
	}()
	if err != nil {
		return 0, fmt.Errorf("when connecting, %w", err)
	}
	// 0x01FE0000 -> byte
	argBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(argBytes, argValue)
	_, err = client.WriteMultipleRegisters(argAddress, 2, argBytes)
	if err != nil {
		return 0, fmt.Errorf("writing argument, %w", err)
	}
	// 0x0001
	_, err = client.WriteSingleRegister(cmdAddress, cmdValue)
	if err != nil {
		return 0, fmt.Errorf("writing command: %w", err)
	}

	b, err := client.ReadHoldingRegisters(argAddress, 2)
	if err != nil {
		return 0, fmt.Errorf("reading return value: %w", err)
	}
	reg1 := getFloat(b, 0)
	reg2 := getFloat(b, 1)
	return uint32(reg1)<<16 | uint32(reg2), nil
}

func (c *ModbusConn) Close() error {
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

func toBytes(value uint64, n int) []byte {
	bytes := make([]byte, n)
	rest := value
	for i := n - 1; i >= 0; i-- {
		b := rest % (1 << 8)
		bytes[i] = byte(b)

		rest = (rest - b) >> 8
	}
	return bytes
}
