package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Client struct {
	Name       string
	Connection net.Conn
}

type Clients []Client

func (c *Client) Enter(r Room) {
	fmt.Fprintln(c.Connection, "You are "+c.Name)
	r.Messages <- c.Name + "has arrived"
	r.Enter <- *c
}

func (c *Client) Leave(r Room) {
	r.Leave <- *c
	r.Messages <- c.Name + " has left"
}

func (c *Client) SendMessage(message string, room Room) {
	room.Messages <- message
}

func (c *Client) ReceveMessage(message string) {
	fmt.Fprintln(c.Connection, "You are "+c.Name)
}

type Room struct {
	Clients  Clients
	Messages chan string
	Enter    chan Client
	Leave    chan Client
}

func NewRoom() Room {
	room := Room{}
	room.Messages = make(chan string)
	room.Enter = make(chan Client)
	room.Leave = make(chan Client)
	go room.Broadcaster()
	return room
}

func (r *Room) Broadcaster() {
	for {
		select {
		case msg := <-r.Messages:
			for _, cli := range r.Clients {
				fmt.Fprintln(cli.Connection, msg)
			}
		case client := <-r.Enter:
			r.Clients = append(r.Clients, client)
		case client := <-r.Leave:
			var newClients Clients
			for _, c := range r.Clients {
				if c.Name == client.Name {
					continue
				}
				newClients = append(newClients, c)
			}
			r.Clients = newClients
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")

	if err != nil {
		log.Fatal(err)
	}
	room := NewRoom()

	run(listener, func(conn net.Conn) {

		client := Client{Name: conn.RemoteAddr().String(), Connection: conn}
		go client.Enter(room)

		input := bufio.NewScanner(client.Connection)
		for input.Scan() {
			client.SendMessage(client.Name+": "+input.Text(), room)
		}
		go client.Leave(room)
		conn.Close()
	})
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func run(listener net.Listener, handle func(c net.Conn)) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(conn)
	}
}
