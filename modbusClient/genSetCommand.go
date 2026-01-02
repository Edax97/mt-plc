package modbusClient

import (
	"fmt"
)

const CodeAddress = 4209
const ArgumentAddress = 4207

const StartStopCode = 0x01
const ArgumentStart = 0x01FE0000
const ArgumentStop = 0x02FD0000

func GenSetON(c *ModbusConn) error {
	returned, err := c.WriteCommand(CodeAddress, StartStopCode, ArgumentAddress, ArgumentStart)
	if err == nil && returned != 0x000001FF {
		return fmt.Errorf("returned value not expected, %d", returned)
	}
	return err
}

func GenSetOFF(c *ModbusConn) error {
	returned, err := c.WriteCommand(CodeAddress, StartStopCode, ArgumentAddress, ArgumentStop)
	if err == nil && returned != 0x000002FE {
		return fmt.Errorf("returned value not expected, %d", returned)
	}
	return err
}
