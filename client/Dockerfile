FROM alpine:3.8

# Notice the client binanry should be generate by 
#GOOS=linux go build
ADD client /client

ENTRYPOINT ["./client"]
#COMMAND["-nats", "localhost:4222", "-sub", "voice-0"]