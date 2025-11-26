package main

import (
	"context"
	"testing"

	"net"

	pb "github.com/PierreDougnac/Todo-gRPC-Service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// server gRPC
var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, newServer())

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

// dial gRPC
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// ---- TESTS ---- //

func TestCreateTodo(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)

	resp, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{
		Title: "Test Todo",
	})
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}

	if resp.Todo.Title != "Test Todo" {
		t.Errorf("expected title 'Test Todo', got %s", resp.Todo.Title)
	}

	if resp.Todo.Id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGetTodo(t *testing.T) {
	ctx := context.Background()

	// Connexion au serveur gRPC en mémoire
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)

	// D'abord créer un todo
	createResp, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{
		Title: "Todo for Get",
	})
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}

	// Maintenant tester GetTodo
	getResp, err := client.GetTodo(ctx, &pb.GetTodoRequest{
		Id: createResp.Todo.Id,
	})
	if err != nil {
		t.Fatalf("GetTodo failed: %v", err)
	}

	if getResp.Todo.Id != createResp.Todo.Id {
		t.Errorf("expected ID %s, got %s", createResp.Todo.Id, getResp.Todo.Id)
	}
}

func TestListTodos(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)

	// Créer deux todos
	client.CreateTodo(ctx, &pb.CreateTodoRequest{Title: "Todo 1"})
	client.CreateTodo(ctx, &pb.CreateTodoRequest{Title: "Todo 2"})

	// Lister
	listResp, err := client.ListTodos(ctx, &pb.ListTodosRequest{})
	if err != nil {
		t.Fatalf("ListTodos failed: %v", err)
	}

	if len(listResp.Todos) < 2 {
		t.Errorf("expected at least 2 todos, got %d", len(listResp.Todos))
	}
}

func TestDeleteTodo(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)

	// Créer un todo
	createResp, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{
		Title: "To delete",
	})
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}

	// Tester delete
	delResp, err := client.DeleteTodo(ctx, &pb.DeleteTodoRequest{
		Id: createResp.Todo.Id,
	})
	if err != nil {
		t.Fatalf("DeleteTodo failed: %v", err)
	}

	if !delResp.Success {
		t.Errorf("expected deletion success")
	}

	// Vérifier que le todo n'existe plus
	_, err = client.GetTodo(ctx, &pb.GetTodoRequest{
		Id: createResp.Todo.Id,
	})
	if err == nil {
		t.Errorf("expected error when getting deleted todo")
	}
}
