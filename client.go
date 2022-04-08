package main

import (
	//"bufio"
	"fmt"
	"io"

	//"io"
	"net"
	"os"
	"strings"
	//"github.com/howeyc/crc16"
)

const ADDRESS = ":3333"
const SUBSCRIBE_COMMAND = "receive"
const SEND_COMMAND = "send"

func cancelConnection(c net.Conn) {
	c.Write([]byte{0x0A, 0x00})
}

func sendError(c net.Conn, err byte) {
	c.Write([]byte{0x08, 0x01, err})
}

func sendFile(c net.Conn, path string, chanel string) {
	fmt.Println(chanel)
	fmt.Println(path)
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer file.Close()
	splitpath := strings.Split(path, "/")
	name := splitpath[len(splitpath)-1]
	message := append([]byte{0x4, byte(len(name)), byte(len(chanel))}, []byte(name)...)
	message = append(message, []byte(chanel)...)
	c.Write(message)
	buffer := make([]byte, 255)
	l, err := c.Read(buffer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(buffer[:l])
	if l == 0 {
		fmt.Println("Error: Empty message")
		return
	}
	switch buffer[0] {
	case 0x05:
		io.Copy(c, file)
		return
	case 0x08:
		fmt.Print("Error: ")
		switch buffer[2] {
		case 0x01:
			fmt.Println("Channel doesn't exist")
		case 0x02:
			fmt.Println("File in process")
		default:
			fmt.Println("Other")
		}
		return
	default:
		fmt.Println("Comunication Error")
		return
	}
}

func subscribe(c net.Conn, channelName string) bool {
	fmt.Println(channelName)
	channelNameLen := len(channelName)
	message := append([]byte{0x02, byte(channelNameLen)}, []byte(channelName)...)
	c.Write(message)
	var buffer [1024]byte
	l, err := c.Read(buffer[:])
	if err != nil {
		return false
	}
	if l == 0 {
		sendError(c, 0x00)
		cancelConnection(c)
		return false
	}
	switch buffer[0] {
	case 3:
		if l == channelNameLen+2 && buffer[1] == byte(channelNameLen) && channelName == string(buffer[2:channelNameLen+2]) {
			break
		}
		sendError(c, 0x02)
	case 0x08:
		cancelConnection(c)
	case 0x0A:
		return false
	default:
		sendError(c, 0x01)
		cancelConnection(c)
		return false
	}
	l, err = c.Read(buffer[:])
	if err != nil {
		return false
	}
	if l < 3 || buffer[0] != 0x06 {
		sendError(c, 0x02)
		return true
	}
	fileName := string(buffer[2 : buffer[1]+2])
	f, err := os.Create("download/" + fileName)
	if err != nil {
		sendError(c, 0x03)
		return false
	}
	c.Write([]byte{0x07, 0x00})
	lf, err := io.Copy(f, c)
	fmt.Println(lf)
	f.Close()
	if err != nil {
		os.Remove("download/" + fileName)
		return false
	}
	return true
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("No Arguments")
		os.Exit(2)
	}
	address := args[1] + ADDRESS
	c, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	switch args[2] {
	case SUBSCRIBE_COMMAND:
		if len(args) != 4 {
			fmt.Println("Wrong Arguments")
		} else if args[3][0] != '-' {
			fmt.Println("Wrong channel format")
		} else {
			subscribe(c, args[3][1:])
		}
	case SEND_COMMAND:
		if len(args) < 5 {
			fmt.Println("Missing arguments")
		} else if len(args) != 5 {
			fmt.Println("Wrong Arguments")
		} else if args[4][0] != '-' {
			fmt.Println("Wrong Channel Format")
		} else {
			sendFile(c, args[3], args[4][1:])
		}
	default:
		fmt.Println("Invalid Command")
	}
	c.Close()
}
