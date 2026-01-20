package modbusClient

import "time"

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
		time.Sleep(time.Millisecond * 80 * time.Duration(tries*tries-tries+1))
		tries = tries + 1
		failFunc()
	}
}
