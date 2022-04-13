Server:
go run server.go start

Client:
    Subscribe:
    go run client.go [address] receive -[channel]

    Upload
    go run client.go [address] send [filepath] -[channel]
