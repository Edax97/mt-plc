package modbusClient

const CodeAddress = 4209
const ArgumentAddress = 4207
const AutomaticStartStopAddr = 4700

const StartStopCode = 0x01
const ArgumentStart = 0x01FE0000
const ArgumentStop = 0x02FD0000

func GenSetON(c *ModbusConn) error {
	return c.WriteCoil(AutomaticStartStopAddr, true)
}

func GenSetOFF(c *ModbusConn) error {
	return c.WriteCoil(AutomaticStartStopAddr, false)
}

//
//func GenSetON(c *ModbusConn) error {
//
//	returned, err := c.WriteCommand(CodeAddress, StartStopCode, ArgumentAddress, ArgumentStart)
//	if err == nil && returned != 0x000001FF {
//		return fmt.Errorf("returned value not expected, %d", returned)
//	}
//	return err
//}
//
//func GenSetOFF(c *ModbusConn) error {
//	returned, err := c.WriteCommand(CodeAddress, StartStopCode, ArgumentAddress, ArgumentStop)
//	if err == nil && returned != 0x000002FE {
//
//		fmt.Printf("Returned: %d\n", returned)
//		return fmt.Errorf("returned value not expected, %d", returned)
//	}
//	return err
//}
