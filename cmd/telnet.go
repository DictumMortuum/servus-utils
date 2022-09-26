package main

import (
	"github.com/ziutek/telnet"
	"time"
)

const timeout = 20 * time.Second

func expect(t *telnet.Conn, d ...string) error {
	err := t.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	err = t.SkipUntil(d...)
	if err != nil {
		return err
	}

	return nil
}

func sendln(t *telnet.Conn, s string) error {
	err := t.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = '\n'

	_, err = t.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func sendSlowly(t *telnet.Conn, s string) error {
	err := t.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	for _, c := range s {
		_, err = t.Write([]byte(string(c)))
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
