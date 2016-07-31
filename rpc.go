package module

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
)

func RpcCodecClient(port string) rpc.ClientCodec {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return jsonrpc.NewClientCodec(conn)
}

func RpcCodecServer(conn io.ReadWriteCloser) rpc.ServerCodec {
	return jsonrpc.NewServerCodec(conn)
}

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
