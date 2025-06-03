// tests/integration/mesh_api_test.go

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	pb "github.com/eli-nomasec/linkerd2-mcp/internal/gen/pb"
	"google.golang.org/grpc"
)

func TestMeshGraphIncludesDemoApp(t *testing.T) {
	// Get MCP address from env or use default
	addr := os.Getenv("MCP _GRPC_ADDR")
	if addr == "" {
		addr = "localhost:10900"
	}

	// Wait for MCP to be ready (retry for up to 30s)
	var conn *grpc.ClientConn
	var err error
	for i := 0; i < 30; i++ {

		conn, err = grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(1*time.Second))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to connect to MCP gRPC at %s: %v", addr, err)
	}
	defer conn.Close()

	client := pb.NewMeshContextClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetMeshGraph(ctx, &pb.GetMeshGraphRequest{})
	if err != nil {
		t.Fatalf("GetMeshGraph failed: %v", err)
	}

	// The mesh graph is returned as JSON
	type Service struct {
		Name string `json:"name"`
	}
	type Edge struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	}
	type MeshGraph struct {
		Services map[string]Service `json:"services"`
		Edges    []Edge             `json:"edges"`
	}

	var graph MeshGraph
	jsonStr := resp.GetJsonGraph()
	if err := json.Unmarshal([]byte(jsonStr), &graph); err != nil {
		t.Fatalf("failed to unmarshal mesh graph JSON: %v", err)
	}

	// Check for service-a and service-b
	_, foundA := graph.Services["service-a"]
	_, foundB := graph.Services["service-b"]
	if !foundA || !foundB {
		t.Fatalf("expected both service-a and service-b in mesh, got: %+v", graph.Services)
	}

	// Check for an edge from service-a to service-b
	var foundEdge bool
	for _, edge := range graph.Edges {
		if edge.Src == "service-a" && edge.Dst == "service-b" {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Fatalf("expected edge from service-a to service-b, got: %+v", graph.Edges)
	}
}
