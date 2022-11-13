package main

import (
	"chatApp-GRPc/proto"
	"context"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	"log"
	"net"
	"os"
	"sync"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
	//grpcLog.Errorf("Error Happened") //test
}

type Connection struct {
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error //Because we are using goroutines, this type needs to be a channel
}

//*******************************************************************************
//'Server' type implements 'BroadcastServer' interface methods from 'service.proto'
type Server struct {
	ConnectionsPool []*Connection //slice of pointers to various Connections
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	//Creating new Connection for the Client
	conn := &Connection{
		stream: stream, //getting from function parameter
		id:     pconn.User.Id,
		active: true,             //true by default
		error:  make(chan error), //new channel
	}
	//Add newly created connection to the Connections Pool on the Server
	s.ConnectionsPool = append(s.ConnectionsPool, conn)

	return <-conn.error //get error from a channel and return it
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) { //returns tuple
	waitGroup := sync.WaitGroup{} //Waits for the  "collection of goroutines" to finish
	done := make(chan int)        //Use it to know when all our goroutines are finished

	for _, conn := range s.ConnectionsPool {
		waitGroup.Add(1) //Incrementing WaitGroup Counter

		//for each connection in the Pool ,spawn a goroutine
		go func(msg *proto.Message, conn *Connection) {
			defer waitGroup.Done() //When goroutine will finish,A Wait Group's Counter will be decremented

			//Here we are passing Messages to the Client
			if conn.active {
				//If Connection is Active send back Message to the Client
				//which is attached to this Connection
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending message to: ", conn.stream)

				if err != nil {
					grpcLog.Errorf("Error with Stream: %s - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err //write error to the Connection's error channel
				}
			}
		}(msg, conn)
	}

	//Create and Call a Goroutine
	go func() {
		//Waits until all goroutines from the wait group Exit
		waitGroup.Wait() //(blocks until group Counter is Zero ).
		close(done)
	}()
	<-done                     //empty done channel before return from the function
	return &proto.Close{}, nil //return expected tuple
}

//*******************************************************************************

func main() {
	var connections []*Connection
	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error creating the Server %v", err)
	}
	grpcLog.Info("Starting server at port :8080")
	//Register BroadcastServer on GRPc Server and passing to it our actual Server type
	proto.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)

}
