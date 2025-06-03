// cmd/mcp-server/main.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"google.golang.org/grpc"

	pb "github.com/eli-nomasec/linkerd2-mcp/internal/gen/pb"
	"github.com/eli-nomasec/linkerd2-mcp/internal/graph"
	redisutil "github.com/eli-nomasec/linkerd2-mcp/internal/redis"
)

type server struct {
	pb.UnimplementedMeshContextServer
	mesh *graph.MeshGraph
}

func (s *server) GetMeshGraph(ctx context.Context, req *pb.GetMeshGraphRequest) (*pb.GetMeshGraphResponse, error) {
	data, err := json.Marshal(s.mesh)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mesh graph: %w", err)
	}
	return &pb.GetMeshGraphResponse{JsonGraph: string(data)}, nil
}

// ApplyAuthorizationPolicy: mutate mesh graph and publish delta to Redis
func (s *server) ApplyAuthorizationPolicy(ctx context.Context, req *pb.ApplyAuthorizationPolicyRequest) (*pb.ApplyAuthorizationPolicyResponse, error) {
	fmt.Printf("Received ApplyAuthorizationPolicy: ns=%s name=%s\n", req.Namespace, req.Name)

	// Parse JSON spec into map
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(req.JsonSpec), &spec); err != nil {
		return &pb.ApplyAuthorizationPolicyResponse{
			Accepted: false,
			Message:  fmt.Sprintf("Invalid JSON spec: %v", err),
		}, nil
	}

	// Update mesh graph in memory
	policyKey := req.Namespace + "/" + req.Name
	s.mesh.AuthPolicies[policyKey] = graph.AuthPolicy{
		Name: req.Name,
		Spec: spec,
	}

	// Serialize delta (currently publishes full mesh graph)
	delta, err := json.Marshal(s.mesh)
	if err != nil {
		return &pb.ApplyAuthorizationPolicyResponse{
			Accepted: false,
			Message:  fmt.Sprintf("Failed to marshal mesh graph: %v", err),
		}, nil
	}

	// Publish delta to Redis
	redis := redisutil.NewRedisClient("localhost:6379")
	if err := redis.PublishMeshDelta(ctx, delta); err != nil {
		return &pb.ApplyAuthorizationPolicyResponse{
			Accepted: false,
			Message:  fmt.Sprintf("Failed to publish mesh delta: %v", err),
		}, nil
	}

	fmt.Printf("Applied and published policy %s\n", policyKey)
	return &pb.ApplyAuthorizationPolicyResponse{
		Accepted: true,
		Message:  "Policy applied and published",
	}, nil
}

func main() {
	fmt.Println("Starting MCP Server...")

	// Initialize mesh graph
	mesh := &graph.MeshGraph{
		Services:     make(map[string]graph.Service),
		Edges:        []graph.Edge{},
		AuthPolicies: make(map[string]graph.AuthPolicy),
	}

	// Initialize Redis client (address would be configurable)
	redis := redisutil.NewRedisClient("localhost:6379")

	// Hydrate mesh from Redis snapshot
	snapshot, err := redis.GetMeshSnapshot(context.Background())
	if err == nil && len(snapshot) > 0 {
		if err := json.Unmarshal(snapshot, mesh); err != nil {
			fmt.Printf("Failed to unmarshal mesh snapshot from Redis: %v\n", err)
		} else {
			fmt.Println("Hydrated mesh graph from Redis snapshot")
		}
	} else {
		fmt.Println("No mesh snapshot found in Redis, starting with empty mesh graph")
	}

	// Subscribe to mesh:delta channel for live updates
	go func() {
		err := redis.SubscribeMeshDelta(context.Background(), func(msg []byte) {
			var patch graph.MeshGraph
			if err := json.Unmarshal(msg, &patch); err != nil {
				fmt.Printf("Failed to unmarshal mesh delta: %v\n", err)
				return
			}
			// Replace mesh with patch (future: merge/patch for efficiency)
			*mesh = patch
			fmt.Println("Applied mesh delta from Redis")
		})
		if err != nil {
			fmt.Printf("Error subscribing to mesh:delta: %v\n", err)
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", ":10900")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMeshContextServer(grpcServer, &server{mesh: mesh})
	fmt.Println("MCP Server listening on :10900")
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
