package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/PierreDougnac/Todo-gRPC-Service/proto"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [args]")
		fmt.Println("Commands: add <title>, list, get <id>, delete <id>")
		return
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)
	cmd := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	switch cmd {
	case "add":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a title for the todo")
		}
		title := os.Args[2]
		resp, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{Title: title})
		if err != nil {
			log.Fatalf("CreateTodo failed: %v", err)
		}
		fmt.Printf("Added todo: %v\n", resp.Todo)
	case "list":
		resp, err := client.ListTodos(ctx, &pb.ListTodosRequest{})
		if err != nil {
			log.Fatalf("ListTodos failed: %v", err)
		}
		fmt.Println("Todos:")
		for _, t := range resp.Todos {
			fmt.Printf("- ID: %s, Title: %s, Completed: %v\n", t.Id, t.Title, t.Completed)
		}
	case "get":
		if len(os.Args) < 3 {
			log.Fatal("Please provide the ID of the todo")
		}
		id := os.Args[2]
		resp, err := client.GetTodo(ctx, &pb.GetTodoRequest{Id: id})
		if err != nil {
			log.Fatalf("GetTodo failed: %v", err)
		}
		fmt.Printf("Todo: %v\n", resp.Todo)
	case "delete":
		if len(os.Args) < 3 {
			log.Fatal("Please provide the ID of the todo")
		}
		id := os.Args[2]
		resp, err := client.DeleteTodo(ctx, &pb.DeleteTodoRequest{Id: id})
		if err != nil {
			log.Fatalf("DeleteTodo failed: %v", err)
		}
		if resp.Success {
			fmt.Println("Todo deleted successfully")
		} else {
			fmt.Println("Todo not found")
		}
	default:
		fmt.Println("Unknown command:", cmd)
	}
}
