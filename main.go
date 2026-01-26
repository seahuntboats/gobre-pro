package main

import (
	"context"
	"fmt"
	server "main/server"
)

func main() {
	fmt.Println("Starting the server...")
	ctx := context.Background()
	server.StartServer(ctx)
}
