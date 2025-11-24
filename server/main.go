package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/PierreDougnac/Todo-gRPC-Service/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTodoServiceServer
	mu     sync.Mutex
	todos  map[string]*pb.Todo
	nextID int
}

func newServer() *server {
	return &server{
		todos:  make(map[string]*pb.Todo),
		nextID: 1,
	}
}

func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.CreateTodoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("%d", s.nextID)
	s.nextID++

	todo := &pb.Todo{
		Id:        id,
		Title:     req.Title,
		Completed: false,
	}
	s.todos[id] = todo

	return &pb.CreateTodoResponse{Todo: todo}, nil
}

func (s *server) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.GetTodoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[req.Id]
	if !ok {
		return nil, fmt.Errorf("todo with id %s not found", req.Id)
	}
	return &pb.GetTodoResponse{Todo: todo}, nil
}

func (s *server) ListTodos(ctx context.Context, req *pb.ListTodosRequest) (*pb.ListTodosResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todos := make([]*pb.Todo, 0, len(s.todos))
	for _, t := range s.todos {
		todos = append(todos, t)
	}
	return &pb.ListTodosResponse{Todos: todos}, nil
}

func (s *server) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.todos[req.Id]; !ok {
		return &pb.DeleteTodoResponse{Success: false}, nil
	}
	delete(s.todos, req.Id)
	return &pb.DeleteTodoResponse{Success: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, newServer())

	log.Println("Todo gRPC server running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
