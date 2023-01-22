# Small Linux container which only have a Golang compiler inside of it
FROM golang:alpine as build-env
#Define Environment variable for this container
ENV GO111MODULE=on
#Run Package Manager insid econtainer,UPDATE sources, ADD 'bash' shell
#ADD 'ca-certificates' for SSL , ADD 'git' cmd tool, ADD 'gcc' and 'g++' compilers
#ADD 'libc-dev' developer libraries for C/C++
RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev
#Setup Derectories structures inside our container
RUN mkdir /go-app-docker_example
RUN mkdir -p /go-app-docker_example/proto
#Setup our Work Directory as '/go-app-docker_example'
WORKDIR /go-app-docker_example

#Copy our Protobuf files and Server's main file to our Docker Container
COPY ./proto/service.pb.go /go-app-docker_example/proto
COPY ./proto/service_grpc.pb.go /go-app-docker_example/proto
COPY ./main.go /go-app-docker_example

#Copy 'go.mod' and  'go.sum' files to container define the Dependencies for our Go Project
COPY go.mod .
COPY go.sum .

#GET Dependencies for our Go Project
RUN  go mod download

#BUILD our Go  Chat Server
RUN go build -o go-app-docker_example .

#Specify the command we want to EXECUTE when we RUN our Docker Container
CMD ./go-app-docker_example