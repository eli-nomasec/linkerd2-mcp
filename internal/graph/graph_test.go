// internal/graph/graph_test.go

package graph

import (
	"testing"
)

func TestMeshGraph_AddService(t *testing.T) {
	g := MeshGraph{
		Services:     make(map[string]Service),
		Edges:        []Edge{},
		AuthPolicies: make(map[string]AuthPolicy),
	}

	svc := Service{
		Name:      "demo-app",
		Namespace: "default",
		Meshed:    true,
	}

	g.Services[svc.Name] = svc

	if len(g.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(g.Services))
	}
	if g.Services["demo-app"].Name != "demo-app" {
		t.Errorf("expected service name 'demo-app', got %s", g.Services["demo-app"].Name)
	}
}
