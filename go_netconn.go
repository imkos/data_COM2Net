// go_netconn
package main

import (
	"fmt"
	"net"
)

type Simple_TcpConn struct {
	net.Conn
}

func NewTcpConn(s_host string) (*Simple_TcpConn, error) {
	cli_conn, e2 := net.Dial("tcp", fmt.Sprintf("%s:9100", s_host))
	if e2 != nil {
		return nil, e2
	}
	return &Simple_TcpConn{cli_conn}, nil
}
