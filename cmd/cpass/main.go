package main

import (
	"flag"
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/koltyakov/gosip/cpass"
)

func main() {

	var masterKey string
	var rawSecret string
	var mode string

	flag.StringVar(&masterKey, "master", "", "Master key")
	flag.StringVar(&rawSecret, "secret", "", "Raw secret string")
	flag.StringVar(&mode, "mode", "encode", "Mode: encode/decode")
	flag.Parse()

	crypt := cpass.Cpass(masterKey)

	if rawSecret == "" && mode == "encode" {
		fmt.Printf("Password to encode: ")
		pass, _ := gopass.GetPasswdMasked()
		secret, _ := crypt.Encode(string(pass))
		fmt.Println(secret)
		return
	}

	if rawSecret == "" && mode == "decode" {
		fmt.Printf("Hash to decode: ")
		hashStr, _ := gopass.GetPasswdMasked()
		secret, _ := crypt.Decode(string(hashStr))
		fmt.Println(secret)
		return
	}

	if rawSecret != "" && mode == "encode" {
		secret, _ := crypt.Encode(rawSecret)
		fmt.Println(secret)
		return
	}

	if rawSecret != "" && mode == "decode" {
		secret, _ := crypt.Decode(rawSecret)
		fmt.Println(secret)
		return
	}

}
