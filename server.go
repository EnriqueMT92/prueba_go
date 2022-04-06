package main

import (
	//"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"

	//"strconv"
	//"strings"
	"time"
	//"github.com/howeyc/crc16"
)

const PORT = ":3333"
const START_COMMAND = "start"

var channels map[string][]net.Conn = make(map[string][]net.Conn)
var fileRecieved map[string]int = map[string]int{}

func main() {
	args := os.Args
	if len(args) != 2 || args[1] != START_COMMAND {
		fmt.Println("Command invalid")
		os.Exit(1)
	}
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer l.Close()
	fmt.Println("Server started port" + PORT)
	rand.Seed(time.Now().Unix())
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		go handleRequest(c)
	}
}
func removeFromChannel(c net.Conn, name string) {
	l := len(channels[name]) - 1
	if l == 1 {
		delete(channels, name)
		return
	}
	for index, val := range channels[name] {
		if val == c {
			channels[name][index] = channels[name][l]
			channels[name][l] = nil
			channels[name] = channels[name][:l]
			return
		}
	}
}
func addToChannel(c net.Conn, name string) {
	if val, found := channels[name]; found {
		channels[name] = append(val, c)
	} else {
		channels[name] = []net.Conn{c}
	}
	fmt.Printf("channels: %v\n", channels)
	c.Write(append([]byte{0x03, byte(len(name))}, []byte(name)...))
	for {
		var buffer [1024]byte
		l, err := c.Read(buffer[:])
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		fmt.Println(buffer[:l])
		if l > 1 && buffer[0] == 0x07 {
			f, err := os.Open("TempFolder/" + name)
			if err != nil {
				break
			}
			fmt.Println("mdkamdamkdmamdakm")
			lf, err := io.Copy(c, f)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(lf)
			f.Close()
			fileRecieved[name]--
			if fileRecieved[name] == 0 {
				delete(fileRecieved, name)
				os.Remove("TempFolder/" + name)
			}
		}
	}
	removeFromChannel(c, name)
}

func recieveFile(c net.Conn, fileName string, channelName string) {
	if channels[channelName] == nil {
		c.Write([]byte{0x08, 0x01, 0x01})
		return
	} else if fileRecieved[channelName] != 0 {
		c.Write([]byte{0x08, 0x01, 0x02})
	}
	f, err := os.Create("TempFolder/" + channelName)
	if err != nil {
		c.Write([]byte{0x08, 0x01, 0x03})
	}
	c.Write([]byte{0x05, 0x01, 0x00})
	_, err = io.Copy(f, c)
	if err != nil {
		return
	}
	f.Close()
	fileRecieved[channelName] = 0
	message := append([]byte{0x06, byte(len(fileName))}, []byte(fileName)...)
	for _, con := range channels[channelName] {
		con.Write(message)
		fileRecieved[channelName]++
	}
}
func handleRequest(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	var buffer [1024]byte
	len, err := c.Read(buffer[:])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if len > 0 {
		switch buffer[0] {
		case 2:
			name := string(buffer[2 : buffer[1]+2])
			addToChannel(c, name)
			break
		case 4:
			name := string(buffer[3 : buffer[1]+3])
			channel := string(buffer[3+buffer[1] : 3+buffer[1]+buffer[2]])
			recieveFile(c, name, channel)
			break
		default:
			fmt.Println(buffer)
		}
	}
	/*data, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Println("Error")
		fmt.Println(err.Error())
		return
	}
	temp := strings.TrimSpace(string(data))
	fmt.Println(temp)
	if temp == "stop" {
		fmt.Printf("Stop Serving %s\n", c.RemoteAddr().String())
		c.Write([]byte("stop"))
		break
	}
	result := strconv.Itoa(random()) + "\n\r"
	c.Write([]byte(result))*/
	c.Close()
}
