package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Server struct {
	IP                 string
	Port               uint16
	UsernamesByClients map[net.Conn]string
	ClientsByUsernames map[string]net.Conn
	Listener           net.Listener
}

func New(IP string, Port uint16) Server {
	var server Server
	server.IP, server.Port, server.UsernamesByClients, server.ClientsByUsernames = IP, Port, make(map[net.Conn]string), make(map[string]net.Conn)
	return server
}

func (server *Server) listen() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	server.handleError(err)
	server.Listener = listener
}

func (server *Server) Launch() {
	server.listen()
	for {
		client, err := server.Listener.Accept()
		server.handleError(err)
		go server.handleClient(client)
	}
}

func (server *Server) handleClient(client net.Conn) {
	for {
		username, err := server.read(client)
		if err != nil {
			break
		}
		if server.isUsernameUnused(username) {
			server.write(fmt.Sprintf("Your username : %s", username), client)
			server.addClient(client, username)
			for {
				message, err := server.read(client)
				if err != nil {
					break
				}
				server.log(fmt.Sprintf("(%s) %s\n", server.UsernamesByClients[client], message))
				server.broadcastToClients(server.formatMessage(message, client), client, true)
			}
			break
		} else {
			server.write("UsernameAlreadyInUse", client)
		}
	}
}

func (server *Server) isUsernameUnused(username string) bool {
	return server.ClientsByUsernames[username] == nil
}

func (server *Server) addClient(client net.Conn, username string) {
	server.ClientsByUsernames[username] = client
	server.UsernamesByClients[client] = username
	server.log(fmt.Sprintf("%s is connected.", username))
	server.broadcastToClients(fmt.Sprintf("***Here is a new friend. Welcome %s!***", username), client, true)
}

func (server *Server) read(client net.Conn) (string, error) {
	reader := bufio.NewReader(client)

	message, err := reader.ReadString('\n')
	canContinue := server.handleReadingError(err, client)
	if canContinue {
		return strings.TrimSuffix(strings.TrimSuffix(message, "\n"), "\r"), nil // Separated in two trims to handle both \r\n and \n end of line
	}
	return "", err
}

func (server *Server) write(message string, client net.Conn) {
	client.Write([]byte(message))
}

func (server *Server) broadcastToClients(message string, sender net.Conn, avoidSender bool) {
	for client := range server.UsernamesByClients {
		if !avoidSender || client != sender {
			server.write(message, client)
		}
	}
}

func (server *Server) formatMessage(message string, client net.Conn) string {
	return fmt.Sprintf("(%s) %s", server.UsernamesByClients[client], message)
}

func (server *Server) handleReadingError(err error, client net.Conn) bool {
	if err != nil {
		server.removeClient(client)
		return false
	}
	return true
}

func (server *Server) removeClient(client net.Conn) {
	server.log(fmt.Sprintf("%s is diconnected.", server.UsernamesByClients[client]))
	server.broadcastToClients(fmt.Sprintf("%s is diconnected. Goodbye!", server.UsernamesByClients[client]), client, true)
	delete(server.ClientsByUsernames, server.UsernamesByClients[client])
	delete(server.UsernamesByClients, client)
	client.Close()
}

func (server *Server) handleError(err error) {
	if err != nil {
		if server.Listener != nil {
			server.Listener.Close()
		}
		fmt.Println("A problem occured with the connection to a client.")
	}
}

func addTime(message string) string {
	return fmt.Sprintf("[%s] %s\n", time.Now().Format("02/01/2006 at 15:04:05"), strings.TrimSuffix(strings.TrimSuffix(message, "\n"), "\r"))
}

func (server *Server) log(logMessage string) {
	logFile, err := os.OpenFile("chatingo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("Log problem.")
	}
	defer logFile.Close()

	if _, err := logFile.WriteString(addTime(logMessage)); err != nil {
		fmt.Println("Log can't be written")
	}
}
