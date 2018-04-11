/**
 * Author: K.o.s
 * Date: 2017-08-16
 * Email: longw@sctek.com
**/
// go_gbk_utf8
package main

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GBK_Decode(b_pi []byte) ([]byte, error) {
	return ioutil.ReadAll(transform.NewReader(bytes.NewReader(b_pi), simplifiedchinese.GBK.NewDecoder()))
}

func GBK_Encode(b_pi []byte) ([]byte, error) {
	return ioutil.ReadAll(transform.NewReader(bytes.NewReader(b_pi), simplifiedchinese.GBK.NewEncoder()))
}
