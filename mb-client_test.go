package main

import (
	"os/exec"
	"testing"
	"time"
)

type register struct {
	address uint16
	value   bool
}

func getAddressSpace() map[string]register {
	return map[string]register{
		"I1": {
			0,
			false,
		},
		"I2": {
			1,
			true,
		},
		"I3": {
			2,
			true,
		},
		"I4": {
			3,
			false,
		},
		"Q1": {
			8192,
			false,
		},
		"Q2": {
			8193,
			false,
		},
		"Q3": {
			8194,
			false,
		},
	}
}

func runBash(script string) error {
	cmd := exec.Command(script, "")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func Test_getBit(t *testing.T) {
	data := []byte{0b01010001} // bits set at positions 0,4,6 (LSB=bit0)
	cases := []struct {
		idx  uint16
		want bool
	}{
		{0, true},
		{1, false},
		{4, true},
		{6, true},
		{7, false},
		{8, false}, // out of range
	}
	for _, c := range cases {
		if got := getBit(data, c.idx); got != c.want {
			t.Errorf("getBit idx=%d got %v want %v", c.idx, got, c.want)
		}
	}
}

func TestModbusConn_ReadDigital(t *testing.T) {

	addresses := getAddressSpace()

	err := runBash("./start-server.sh")
	defer func() {
		err = runBash("./kill-server.sh")
		if err != nil {
			t.Fatalf("could not kill server, %v", err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}

	con, err := NewModbusConn("0.0.0.0:5020", time.Second)
	defer func() {
		if con == nil {
			return
		}
		_ = con.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	type testCase struct {
		typ  rune
		ad   []uint16
		want []bool
	}

	cases := map[string]testCase{
		"single input": {'I',
			[]uint16{
				addresses["I1"].address,
			},
			[]bool{
				addresses["I1"].value,
			}},
		"contiguous inputs": {'I',
			[]uint16{
				addresses["I2"].address, addresses["I3"].address,
			},
			[]bool{
				addresses["I2"].value, addresses["I3"].value,
			},
		},
		"discontiguous inputs": {'I',
			[]uint16{
				addresses["I1"].address, addresses["I4"].address,
			},
			[]bool{
				addresses["I1"].value, addresses["I4"].value,
			},
		},
		"single output": {'Q',
			[]uint16{
				addresses["Q1"].address,
			}, []bool{
				addresses["Q1"].value,
			}},
		"contiguous outputs": {'Q',
			[]uint16{
				addresses["Q2"].address, addresses["Q3"].address,
			},
			[]bool{
				addresses["Q2"].value, addresses["Q3"].value,
			},
		},
		"discontiguous outputs": {'Q',
			[]uint16{
				addresses["Q1"].address, addresses["Q3"].address,
			},
			[]bool{
				addresses["Q1"].value, addresses["Q3"].value,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var readFunc func(a []uint16) ([]bool, error)
			if tc.typ == 'I' {
				readFunc = con.ReadInputs
			} else {
				readFunc = con.ReadCoils
			}
			got, err := readFunc(tc.ad)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Errorf("want: %v\n, got: %v", tc.want, got)

			}
			for j, w := range tc.want {
				if w != got[j] {
					t.Errorf("error readDigital\nwant:%v\ngot:%v", tc.want, got)
					return
				}
			}

		})
	}

}

func TestModbusConn_WriteCoil(t *testing.T) {

	addresses := getAddressSpace()

	err := runBash("./start-server.sh")
	defer func() {
		err = runBash("./kill-server.sh")
		if err != nil {
			t.Fatalf("could not kill server, %v", err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}

	con, err := NewModbusConn("0.0.0.0:5020", time.Second)
	defer func() {
		if con == nil {
			return
		}
		_ = con.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		ad      uint16
		set     bool
		success bool
	}

	cases := map[string]testCase{
		"single coil on": {
			addresses["Q1"].address,
			true,
			true,
		},
		"single coil off": {
			addresses["Q1"].address,
			false,
			true,
		},
		"single coil off Q3": {
			addresses["Q3"].address,
			false,
			true,
		},
		"single coil fail": {
			8888,
			true,
			false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := con.WriteCoil(tc.ad, tc.set)
			if !tc.success {
				if err == nil {
					t.Errorf("should failed when writing at %d", tc.ad)
				}
				return
			}
			if err != nil {
				t.Errorf("wanted to write %t to %d, got %v",
					tc.set, tc.ad, err)
				return
			}

			got, err := con.ReadCoils([]uint16{tc.ad})
			if err != nil {
				t.Error(err)
				return
			}

			if got[0] != tc.set {
				t.Errorf("wanted: %t, got %t (at address %d)", tc.set, got[0], tc.ad)
			}

		})
	}

}
