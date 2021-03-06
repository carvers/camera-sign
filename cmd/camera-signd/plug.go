package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const maxResponseSize = 4096

type Hs1xxPlug struct {
	IPAddress string
}

func (p *Hs1xxPlug) TurnOn() error {
	json := `{"system":{"set_relay_state":{"state":1}}}`
	data := encrypt(json)
	_, err := send(p.IPAddress, data)
	return err
}

func (p *Hs1xxPlug) TurnOff() error {
	json := `{"system":{"set_relay_state":{"state":0}}}`
	data := encrypt(json)
	_, err := send(p.IPAddress, data)
	return err
}

func (p *Hs1xxPlug) SystemInfo() (string, error) {
	json := `{"system":{"get_sysinfo":{}}}`
	data := encrypt(json)
	reading, err := send(p.IPAddress, data)
	if err != nil {
		return "", err
	}
	results := decrypt(reading)
	return results, nil
}

func (p *Hs1xxPlug) MeterInfo() (string, error) {
	json := `{"system":{"get_sysinfo":{}}, "emeter":{"get_realtime":{},"get_vgain_igain":{}}}`
	data := encrypt(json)
	reading, err := send(p.IPAddress, data)
	if err != nil {
		return "", nil
	}
	return decrypt(reading), nil
}

func (p *Hs1xxPlug) DailyStats(month int, year int) (string, error) {
	json := fmt.Sprintf(`{"emeter":{"get_daystat":{"month":%d,"year":%d}}}`, month, year)
	data := encrypt(json)
	reading, err := send(p.IPAddress, data)
	if err != nil {
		return "", err
	}
	return decrypt(reading), nil
}

func encrypt(plaintext string) []byte {
	n := len(plaintext)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(n))
	ciphertext := []byte(buf.Bytes())

	key := byte(0xAB)
	payload := make([]byte, n)
	for i := 0; i < n; i++ {
		payload[i] = plaintext[i] ^ key
		key = payload[i]
	}

	for i := 0; i < len(payload); i++ {
		ciphertext = append(ciphertext, payload[i])
	}

	return ciphertext
}

func decrypt(ciphertext []byte) string {
	n := len(ciphertext)
	key := byte(0xAB)
	var nextKey byte
	for i := 0; i < n; i++ {
		nextKey = ciphertext[i]
		ciphertext[i] = ciphertext[i] ^ key
		key = nextKey
	}
	return string(ciphertext)
}

func send(ip string, payload []byte) ([]byte, error) {
	// 10 second timeout
	conn, err := net.DialTimeout("tcp", ip+":9999", time.Duration(10)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to plug: %w", err)
	}
	defer conn.Close()
	_, err = conn.Write(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot write data to plug: %w", err)
	}
	data := make([]byte, 4096)
	num, err := conn.Read(data)
	data = data[4:num]
	if err != nil {
		return data, err
	}
	return data, nil
}
