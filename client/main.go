package main

import (
	"bufio"
	"chatApp-GRPc/proto"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"sync"
	"time"
)

var client proto.BroadcastClient
var wait *sync.WaitGroup

//Initializing Wait Group
func init() {
	wait = &sync.WaitGroup{}
}

//----------Defining 'connect()' function for our BroadcastClient-----------------
func connect(user *proto.User) error {
	var streamError error

	stream, err := client.CreateStream(context.Background(), &proto.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("connection failed %v", err)
	}
	wait.Add(1)
	go func(str proto.Broadcast_CreateStreamClient) {
		defer wait.Done()
		for {
			msg, err := str.Recv() //Waiting for messages from our  Server
			if err != nil {
				streamError = fmt.Errorf("Error reading message: %v", err)
				break
			}
			//Printing out the Messages
			fmt.Printf("%v : %s\n", msg.Id, msg.Content)
		}

	}(stream)

	return streamError
}

//-----------------------------------------------------------------------------------

func main() {
	//******  Get User Info ******************
	timestamp := time.Now() //Get Time Stamp
	done := make(chan int)  //Setup integer channel
	name := flag.String("N", "Anonymous", "The name of the User")
	flag.Parse()
	id := sha256.Sum256([]byte(timestamp.String() + *name)) //Generate an ID for the User

	//******** Connecting to the Server" ***************
	//We ae not using HTTPS in our Example
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Couldnt Connect to the Service: %v", err)
	}
	//****Create new BroadCastClient on the Connection we made previously******
	client = proto.NewBroadcastClient(conn)
	user := &proto.User{
		Id:   hex.EncodeToString(id[:]),
		Name: *name,
	}
	//***Connecting to our BroadCastClient with our User Credentials
	connect(user)

	wait.Add(1) //Incrementing our Wait Group because we are creating another Go Routine

	//********** Creating Message Handling Loop(goroutine) for our User
	go func() {
		defer wait.Done()
		//Scanning the Input from the User(Standard Input)
		scanner := bufio.NewScanner(os.Stdin)
		//Loops until there no more Input
		for scanner.Scan() {
			//For each of the Loops we're creating a new Message
			msg := &proto.Message{
				Id:        user.Id,
				Content:   scanner.Text(),
				Timestamp: timestamp.String(),
			}
			//Sending Message to the Server
			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error Sending Message:%v", err)
				break
			}
		}
	}()
	//Another goroutine waiting for our WaitGroup to decrement all the way down
	go func() {
		wait.Wait()
		close(done)
	}()

	//Waiting until the 'wait' channel sends back something
	//It will not end until 'done' will send back some data
	<-done

}
