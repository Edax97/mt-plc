package modbusClient

type tryFuncT func() ([]byte, error)
type failFuncT func()

func tryNTimes(tryFunc tryFuncT, failFunc failFuncT, n int) ([]byte, error) {
	tries := 1
	for {
		b, err := tryFunc()
		if err == nil {
			return b, err
		}
		if tries == n {
			return nil, err
		}
		tries = tries + 1
		failFunc()
	}
}
