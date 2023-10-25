package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/ITU-Distributed-System-2023/proto"
	"google.golang.org/grpc"
)

var lamportClock int64

type server struct {
	pb.UnimplementedChittyChatServer // Necessary
	mu                               sync.Mutex
	clients                          map[int64]chan *pb.ChatMessage
	clientStreams                    map[int64]pb.ChittyChat_ChatServer // 新增字段，用于保存客户端流
}

func (s *server) Chat(stream pb.ChittyChat_ChatServer) error {
	msg, err := stream.Recv()
	if err != nil {
		log.Printf("Error receiving first message: %v", err)
		return err
	}
	clientID := msg.ClientId
	clientMsgChan := make(chan *pb.ChatMessage)
	s.mu.Lock()
	s.clients[clientID] = clientMsgChan
	s.clientStreams[clientID] = stream // 将客户端的流存储在映射中
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		delete(s.clientStreams, clientID) // 客户端断开连接时从映射中删除流
		s.mu.Unlock()

		// Notify all clients about the leaving participant
		lamportClock++
		leaveMessage := &pb.ChatMessage{
			Content:     fmt.Sprintf("Participant %d left Chitty-Chat ", clientID),
			MessageType: "Info",
		}
		s.mu.Lock()
		for _, clientStream := range s.clientStreams {
			if err := clientStream.Send(leaveMessage); err != nil {
				log.Printf("Error sending leave message to a client: %v", err)
			}
		}
		s.mu.Unlock()
	}()

	// Notify all clients about the new participant
	joinMessage := &pb.ChatMessage{
		Content:     fmt.Sprintf("Participant %d joined Chitty-Chat ", clientID),
		MessageType: "Info",
	}
	s.mu.Lock()
	for _, clientStream := range s.clientStreams {
		if err := clientStream.Send(joinMessage); err != nil {
			log.Printf("Error sending join message to a client: %v", err)
		}
	}
	s.mu.Unlock()

	// Send the list of connected client ports to the new client
	s.mu.Lock()
	var clientPorts []int64
	for id := range s.clients {
		clientPorts = append(clientPorts, id)
	}
	s.mu.Unlock()
	lamportClock++
	responseMsg := &pb.ChatMessage{
		Content:     "Connected clients: " + fmt.Sprint(clientPorts),
		MessageType: "Info",
	}
	if err := stream.Send(responseMsg); err != nil {
		log.Printf("Error sending client list to new client: %v", err)
	}

	// Handle client messages
	func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				break
			}

			if msg.Clock+1 > lamportClock {
				lamportClock = msg.Clock + 1
			} else {
				lamportClock++
			}
			fmt.Println(msg)

			// Broadcast the message to all connected clients
			s.mu.Lock()
			for id, clientStream := range s.clientStreams {
				if id != clientID { // Exclude the sender
					msg.Clock = lamportClock
					if err := clientStream.Send(msg); err != nil {
						log.Printf("Error sending message to client %d: %v", id, err)
					}
					lamportClock++
				}
			}
			s.mu.Unlock()
		}
	}()

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051") // Listen on port 50051
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterChittyChatServer(s, &server{
		clients:       make(map[int64]chan *pb.ChatMessage),
		clientStreams: make(map[int64]pb.ChittyChat_ChatServer),
	})
	fmt.Println("Server is listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
