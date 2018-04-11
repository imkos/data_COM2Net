/**
 * Author: K.o.s
 * Date: 2017-08-18
 * Email: longw@sctek.com
**/
// go_serial
package main

import (
	"github.com/tarm/serial"
)

type Simple_serial struct {
	scfg *serial.Config
	*serial.Port
}

func NewSimpleserial(s_name string, i_baud int) (*Simple_serial, error) {
	ss := &Simple_serial{
		scfg: &serial.Config{
			Name: s_name,
			Baud: i_baud,
		},
	}
	sp, err := serial.OpenPort(ss.scfg)
	if err != nil {
		return nil, err
	}
	ss.Port = sp
	return ss, nil
}
