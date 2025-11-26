package main

import (
	"testing"
	"time"
)

func setAdrr() (a AddrMap) {
	a = AddrMap{
		I1: uint16(1024),
		I2: uint16(1025),
		Q1: uint16(1064),
		Q2: uint16(1065),
		M1: uint16(1066),
		M2: uint16(1067),
	}
	return
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

func TestReadDigInput(t *testing.T) {
	con, err := NewModbusConn("0.0.0.0:5020", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer con.Close()

	a := setAdrr()
	got, err := con.ReadInputs([]uint16{a.I1, a.I2})

	if err != nil {
		t.Fatalf("ReadInputs error: %v", err)
	}

	want := []bool{true, true}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReadInputs[%d]=%v want %v", i, got[i], want[i])
		}
	}
}

func TestReadCoil(t *testing.T) {
	con, err := NewModbusConn("0.0.0.0:5020", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer con.Close()

	a := setAdrr()
	got, err := con.ReadCoils([]uint16{a.Q2})
	if err != nil {
		t.Fatalf("ReadCoil error: %v", err)
	}
	want := []bool{false}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReadCoils[%d]=%v want %v", i, got[i], want[i])
		}
	}
}

// command coil=1
func TestWriteCoil(t *testing.T) {
	con, err := NewModbusConn("0.0.0.0:5020", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer con.Close()

	a := setAdrr()
	err = con.WriteCoil(a.Q1, true)
	if err != nil {
		t.Fatalf("WriteCoil error: %v", err)
	}

	got, err := con.ReadCoils([]uint16{a.Q1})
	want := []bool{true}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReadCoils[%d]=%v want %v", i, got[i], want[i])
		}
	}
}
