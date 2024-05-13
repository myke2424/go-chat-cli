package main

import "fmt"
import "os"
import "strings"
import "bufio"
import "net"

func Connect(host string, port int) net.Conn {
  address := fmt.Sprintf("%s:%d", host, port)
  fmt.Printf("Connecting to server [%s]\n", address)
  conn, err := net.Dial("tcp", address)
  if err != nil {
    panic("Failed to connect to server!")
  }
  fmt.Println("Connected")
  return conn
}

func SendMessage(msg []byte, conn net.Conn) {
 fmt.Printf("Sending message to server: [%s]\n", string(msg))
 _, err := conn.Write(msg) 
 if err != nil {
   panic("Failed to send message to server")
 } 
 fmt.Println("Message sent")
}

func main() {
  conn := Connect("localhost", 8080)
  defer conn.Close()
  scanner := bufio.NewScanner(os.Stdin)

  for {
    scanner.Scan()
    msg := scanner.Text()

    if strings.ToLower(msg) == "exit" {
      fmt.Println("Exiting chat room")
      break
    }
    SendMessage([]byte(msg), conn)

    response, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
      fmt.Println("Error reading response from server:", err)
    }
    fmt.Println("Response from server: ", response)
  }
}
