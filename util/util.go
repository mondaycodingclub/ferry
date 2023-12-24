package util

import (
	"bufio"
	"encoding/json"
	"ferry/api"
	"fmt"
	"io"
	"net"
	"time"
)

const TCP = "tcp"

const DELIMITER byte = '\n'

func Copy(src, dst net.Conn) {
	go func() {
		fmt.Println("copy start", src.RemoteAddr().String(), "->", dst.RemoteAddr().String())
		_, err := io.Copy(src, dst)
		fmt.Println("copy end", src.RemoteAddr().String(), "->", dst.RemoteAddr().String())
		if err != nil {
			return
		}
		defer src.Close()
	}()

	go func() {
		fmt.Println("copy start", dst.RemoteAddr().String(), "->", src.RemoteAddr().String())
		_, err := io.Copy(dst, src)
		fmt.Println("copy end", dst.RemoteAddr().String(), "->", src.RemoteAddr().String())
		if err != nil {
			return
		}
		defer dst.Close()
	}()
}

func Read(conn net.Conn, delim byte) (string, error) {
	reader := bufio.NewReader(conn)
	return reader.ReadString(delim)
}

func Write(conn net.Conn, content string) (int, error) {
	writer := bufio.NewWriter(conn)
	number, err := writer.WriteString(content + "\n")
	if err == nil {
		err = writer.Flush()
	}
	return number, err
}

func KeepAlive(conn net.Conn) {
	for {
		ctrl := api.Ctrl{
			Type: "KeepALive",
		}
		content, err := json.Marshal(ctrl)
		if err != nil {
			return
		}
		_, err = Write(conn, string(content))
		if err != nil {
			return
		}
		time.Sleep(time.Second * 3)
	}
}
