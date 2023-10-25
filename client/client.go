package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ITU-Distributed-System-2023/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	id           int
	portNumber   int
	lamportClock int
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
)

func newEvent(client *Client, messageClock int) {
	// change the clock
	if client.lamportClock < messageClock {
		client.lamportClock = messageClock + 1
	} else {
		client.lamportClock++
	}
}

func newLocalEvent(client *Client) {
	// change the clock
	client.lamportClock++
}

func connectToServer() (proto.ChittyChatClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	} else {
		log.Printf("Connected to the server at port %d\n", *serverPort)
	}
	return proto.NewChittyChatClient(conn), err
}

func joinChat(client *Client) {
	clientConn, err := connectToServer()
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	}

	clientMessageStream, err := clientConn.Chat(context.Background())
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	}

	newLocalEvent(client)
	joinMessage := &proto.ChatMessage{
		Clock:       int64(client.lamportClock),
		ClientId:    int64(client.id),
		Content:     "Join",
		MessageType: "Join",
	}
	if err := clientMessageStream.Send(joinMessage); err != nil {
		log.Fatalf("Could not join: %v", err)
	}

	// Start a goroutine to receive and display messages from the Chitty-Chat
	go func() {
		for {
			serverMessage, err := clientMessageStream.Recv()
			if err != nil {
				log.Fatalf("Failed to receive a message: %v", err)
				return
			}
			// new event
			newEvent(client, int(serverMessage.Clock))
			log.Printf("Received Message : %s, client lamport clock: %d\n", serverMessage.Content, client.lamportClock)
		}
	}()

	// Start a goroutine to send messages to the Chitty-Chat

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		newLocalEvent(client)

		if text == "exit" {
			leaveMessage := &proto.ChatMessage{
				Clock:       int64(client.lamportClock),
				ClientId:    int64(client.id),
				Content:     "Leave",
				MessageType: "Leave",
			}
			clientMessageStream.Send(leaveMessage)

			time.Sleep(1 * time.Second)
			break
		}

		// Send a chat message to the Chitty-Chat
		client.lamportClock++
		chatMessage := &proto.ChatMessage{
			Clock:       int64(client.lamportClock),
			ClientId:    int64(client.id),
			Content:     text,
			MessageType: "Message",
		}
		clientMessageStream.Send(chatMessage)
	}

	os.Exit(0)
}

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	client := &Client{
		id:           *clientPort,
		portNumber:   *clientPort,
		lamportClock: 0,
	}

	// connet to the server

	joinChat(client)

	for {

	}
}
