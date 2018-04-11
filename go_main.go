// go_main
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	bo_hex_out  = false
	bo_gbk_open = false
)

var (
	si_xun = []byte{0x0a, 0x0d, 0x0a, 0x0a}
	ke_mai = []byte{0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0d, 0x0a}
	//
	files_path = filepath.Dir(os.Args[0]) + string(filepath.Separator) + "bb_file"
)

func DoPrinterListen() (ln net.Listener) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	done := make(chan struct{}, 1)
	DoNetListen := func() {
		var err error
		//开启端口监听
		ln, err = net.Listen("tcp", ":9100")
		if err != nil {
			fmt.Println("net.Listen error:", err)
			return
		}
		done <- struct{}{}
	}
	go DoNetListen()
	for {
		select {
		case <-done:
			fmt.Println("PrinterListen Success!")
			return
		case <-ticker.C:
			DoNetListen()
		}
	}
}

func handleMessage(conn net.Conn, ch_d chan<- []byte) {
	msg_buf := new(bytes.Buffer)
	defer func() {
		ch_d <- msg_buf.Bytes()
	}()
	//128kb的缓冲区
	//10kb的缓冲区
	buffer := make([]byte, 1024*10)
	for {
		readLen, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("conn Close!")
				conn.Close()
				return
			}
			fmt.Println(err)
			return
		}
		//log.Println(buffer[:readLen])
		msg_buf.Write(buffer[:readLen])
		if msg_buf.Len() >= 3 {
			buf := msg_buf.Bytes()
			if bytes.Compare(buf[:3], []byte{16, 4, 2}) == 0 {
				log.Println("v1")
				conn.Write([]byte{18})
				msg_buf.Reset()
			} else if bytes.Compare(buf[:3], []byte{29, 114, 1}) == 0 {
				log.Println("v2")
				conn.Write([]byte{0})
				msg_buf.Reset()
			} else if bytes.Compare(buf[len(buf)-3:], []byte{29, 114, 1}) == 0 {
				log.Println("v3")
				conn.Write([]byte{0})
				ch_d <- buf[:len(buf)-3]
				msg_buf.Reset()
			}
		}
	}
}

func tcp_Listen() {
	ch_data := make(chan []byte, 100)
	go func() {
		for {
			select {
			case buf_c := <-ch_data:
				log.Println("buf_c len:", len(buf_c))
				if len(buf_c) == 0 {
					break
				}
				if !bo_hex_out {
					if bo_gbk_open {
						b_src, er := GBK_Decode(buf_c)
						if er != nil {
							fmt.Println("GBK_Decode error", er)
						}
						log.Println("buf_c context:", string(b_src))
					} else {
						f_rev, _ := os.Create(files_path + string(filepath.Separator) + fmt.Sprintf("%d", currentTimeMillis()) + ".bb")
						f_rev.Write(buf_c)
						f_rev.Close()
					}
				} else {
					fmt.Println(hex.Dump(buf_c))
				}
			}
		}
	}()
	//开启打印机9100的tcp监听
	ln := DoPrinterListen()
	defer ln.Close()
	//进入消息处理
	for {
		conn, _ := ln.Accept()
		log.Println("New conn")
		go handleMessage(conn, ch_data)
	}
}

//
type VCOM_token struct {
	b_token []byte
	b_len   int
}

func (v *VCOM_token) IsEnd(b_end []byte) bool {
	return bytes.Compare(b_end[len(b_end)-v.Len():], v.b_token) == 0
}

func (v *VCOM_token) Len() int {
	if v.b_len == 0 {
		v.b_len = len(v.b_token)
	}
	return v.b_len
}

func vcom_Listen_default() {
	ch := make(chan int)
	send_data := func(d []byte) {
		log.Println(d)
		/*
			tcp_conn, err := NewTcpConn("192.168.1.192")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer tcp_conn.Close()
			if _, e0 := tcp_conn.Write(d); e0 != nil {
				fmt.Println("tcp_conn.Write error:", e0)
			}
			time.Sleep(500)
		*/
	}
	sp, e1 := NewSimpleserial("COM3", 115200)
	if e1 != nil {
		log.Println("NewSimpleserial err:", e1)
		return
	}
	go func() {
		//5kb 缓存
		buf := make([]byte, 5*1024)
		for {
			n, err := sp.Read(buf)
			if err != nil {
				fmt.Println(err)
			}
			send_data(buf[:n])
		}
	}()
	ch <- 0
}

//虚拟串口监听
func vcom_Listen(vc *VCOM_token) {
	if vc == nil {
		fmt.Println("VCOM_token is nil")
		return
	}
	ch := make(chan int)
	send_data := func(d *bytes.Buffer) {
		tcp_conn, err := NewTcpConn("192.168.1.192")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer tcp_conn.Close()
		if _, e0 := d.WriteTo(tcp_conn); e0 != nil {
			fmt.Println("tcp_conn.Write error:", e0)
		}
		time.Sleep(500)
	}
	sp, e1 := NewSimpleserial("COM3", 115200)
	if e1 != nil {
		log.Println("NewSimpleserial err:", e1)
		return
	}
	go func() {
		//5kb 缓存
		buf := make([]byte, 5*1024)
		data_buf := new(bytes.Buffer)
		for {
			n, err := sp.Read(buf)
			if err != nil {
				fmt.Println(err)
			}
			data_buf.Write(buf[:n])
			if data_buf.Len() > vc.Len() {
				real_data := data_buf.Bytes()
				if vc.IsEnd(real_data) {
					send_data(data_buf)
				}
			}
		}
	}()
	ch <- 0
}

func o_test() {
	var final = strings.Split("a b c", " ")
	var first = strings.Split("d e f", " ")
	final = append(final, first[2:]...)
	//gofmt -s main.go would change the previous line to
	//final = append(final, first[2:]...)
}

func tcp_client() {
	cli, e1 := NewTcpConn("10.20.60.201")
	if e1 != nil {
		fmt.Println(e1)
		return
	}
	buf := make([]byte, 1024)
	go func() {
		nr, err := cli.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("buf:", buf[:nr])
	}()
	pos_cmd := []byte{0x10, 0x04, 0x02}
	//pos_cmd := []byte{0x1d, 0x72, 0x01}
	//pos_cmd, _ := hex.DecodeString("1d07020101")
	cli.Write(pos_cmd)
	time.Sleep(5 * time.Second)
}

//判断文件或文件夹是否存在
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / 1000000
}

func main() {
	if !Exist(files_path) {
		os.MkdirAll(files_path, os.ModePerm)
	}
	tcp_Listen()
	//vcom_Listen(&VCOM_token{b_token: ke_mai})
	//vcom_Listen_default()
	//tcp_client()
}
