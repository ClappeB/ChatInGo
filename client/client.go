package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MAX_USERNAME_LENGTH = 20
)

var waitGroup sync.WaitGroup

type Client struct {
	IP          string
	PORT        uint16
	Username    string
	Connection  net.Conn
	isConnected bool
}

func New(IP string, PORT uint16) Client {
	var client Client
	client.IP, client.PORT, client.isConnected = IP, PORT, false
	return client
}

func (client *Client) read() (string, error) {
	messageBuffer := make([]byte, 4096)
	length, err := client.Connection.Read(messageBuffer)
	if err != nil {
		fmt.Println("A problem occured with the connection to the server.")
	}
	return string(messageBuffer[:length]), err
}

func (client *Client) askForUnusedUsername() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your username : ")
		username, err := reader.ReadString('\n')
		exitIfError(err)

		usernameLength := len(username)
		if usernameLength <= 0 || usernameLength > MAX_USERNAME_LENGTH {
			fmt.Println("The username you took isn't conform:")
			fmt.Printf("  - Max length: %d\n\n", MAX_USERNAME_LENGTH)
			continue
		}

		if client.isUsernameUnused(username) {
			client.Username = username
			fmt.Printf("Welcome %s. Start chatting by typing a message like \"Hello!\"\n", strings.TrimSuffix(strings.TrimSuffix(username, "\n"), "\r"))
			break
		} else {
			fmt.Println("This username is already taken.")
		}
	}
}

func (client *Client) isUsernameUnused(username string) bool {
	client.Connection.Write([]byte(username))
	response, err := client.read()
	if err != nil {
		client.Connection.Close()
		os.Exit(1)
	}
	return response != "UsernameAlreadyInUse"
}

func (client *Client) connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", client.IP, strconv.Itoa(int(client.PORT))))
	if err != nil {
		if client.Connection != nil {
			client.Connection.Close()
		}
		fmt.Print("Error while connecting. The program will shutdown.")
		time.Sleep(2000)
		os.Exit(1)
	}
	client.Connection, client.isConnected = conn, true
	fmt.Println("Connected to", client.Connection.RemoteAddr())
}

func exitIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func (client *Client) Launch() {
	client.connect()
	client.askForUnusedUsername()
	waitGroup.Add(2)
	go client.receive()
	go client.send()
	waitGroup.Wait()
	fmt.Println("Successfully disconnected.")
}

func (client *Client) send() {
	defer waitGroup.Done()
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		client.Connection.Write([]byte(message))
	}
}

func (client *Client) receive() {
	defer waitGroup.Done()
	for {
		message, err := client.read()
		if err == nil {
			fmt.Println(message)
		} else {
			fmt.Println("Impossible to get message from the server.\nShutting down...")
			time.Sleep(2000)
			os.Exit(1)
		}
	}
}
