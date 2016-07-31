package module

import (
	"net"
	"strconv"
)

func getOpenPort() string {
	for i := 65535; i >= 49152; i-- {
		conn, err := net.Listen("tcp", ":"+strconv.Itoa(i))
		if err == nil {
			conn.Close()
			return strconv.Itoa(i)
		}
	}
	return ""
}
