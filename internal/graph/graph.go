// internal/graph/graph.go

package graph

type Service struct {
	Name      string
	Namespace string
	Meshed    bool
}

type Edge struct {
	Src string
	Dst string
	RPS float64
	TLS bool
}

type AuthPolicy struct {
	Name string
	Spec map[string]interface{}
}

type MeshGraph struct {
	Services     map[string]Service
	Edges        []Edge
	AuthPolicies map[string]AuthPolicy
}
