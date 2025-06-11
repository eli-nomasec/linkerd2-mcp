// cmd/collector/main.go

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eli-nomasec/linkerd2-mcp/internal/graph"
	redisutil "github.com/eli-nomasec/linkerd2-mcp/internal/redis"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"encoding/json"

	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type CollectorConfig struct {
	RedisURL      string
	PrometheusURL string
}

func getConfigFromEnv() CollectorConfig {
	redisURL := os.Getenv("MCP_COLLECTOR_REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	promURL := os.Getenv("MCP_COLLECTOR_PROMETHEUS_URL")
	if promURL == "" {
		promURL = "http://localhost:9090"
	}
	return CollectorConfig{
		RedisURL:      redisURL,
		PrometheusURL: promURL,
	}
}

func main() {
	fmt.Println("Starting MCP Collector...")

	// Set up context that cancels on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load config (env overrides defaults)
	cfg := getConfigFromEnv()
	fmt.Printf("Using Redis URL: %s\n", cfg.RedisURL)
	fmt.Printf("Using Prometheus URL: %s\n", cfg.PrometheusURL)

	// Initialize mesh graph
	mesh := graph.MeshGraph{
		Services:     make(map[string]graph.Service),
		Edges:        []graph.Edge{},
		AuthPolicies: make(map[string]graph.AuthPolicy),
	}

	// Initialize Redis client
	redis := redisutil.NewRedisClient(cfg.RedisURL)

	// Initialize Kubernetes client
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		fmt.Printf("Failed to load kubeconfig: %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create k8s client: %v\n", err)
		os.Exit(1)
	}
	// Initialize dynamic client for CRDs
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create dynamic client: %v\n", err)
		os.Exit(1)
	}

	// Create informer factory
	factory := informers.NewSharedInformerFactory(clientset, 0)

	// Add informer for Service resources
	serviceInformer := factory.Core().V1().Services().Informer()
	// Add informer for Pod resources
	podInformer := factory.Core().V1().Pods().Informer()
	stopCh := make(chan struct{})
	defer close(stopCh)

	go serviceInformer.Run(stopCh)
	go podInformer.Run(stopCh)

	// Add event handlers to update mesh graph on Service add/update/delete
	serviceInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				svc, ok := obj.(*corev1.Service)
				if !ok {
					fmt.Println("Service add: type assertion failed")
					return
				}
				// Detect mesh membership: check if any pod in the service's namespace has the linkerd-proxy container
				meshed := false
				pods, err := clientset.CoreV1().Pods(svc.Namespace).List(context.Background(), metav1.ListOptions{
					LabelSelector: fmt.Sprintf("app=%s", svc.Name),
				})
				if err == nil {
					for _, pod := range pods.Items {
						for _, c := range pod.Spec.Containers {
							if c.Name == "linkerd-proxy" {
								meshed = true
								break
							}
						}
						if meshed {
							break
						}
					}
				}
				mesh.Services[svc.Name] = graph.Service{
					Name:      svc.Name,
					Namespace: svc.Namespace,
					Meshed:    meshed,
				}
				fmt.Printf("Service added: %s/%s\n", svc.Namespace, svc.Name)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				svc, ok := newObj.(*corev1.Service)
				if !ok {
					fmt.Println("Service update: type assertion failed")
					return
				}
				mesh.Services[svc.Name] = graph.Service{
					Name:      svc.Name,
					Namespace: svc.Namespace,
					Meshed:    false, // TODO: Detect mesh membership
				}
				fmt.Printf("Service updated: %s/%s\n", svc.Namespace, svc.Name)
			},
			DeleteFunc: func(obj interface{}) {
				svc, ok := obj.(*corev1.Service)
				if !ok {
					fmt.Println("Service delete: type assertion failed")
					return
				}
				delete(mesh.Services, svc.Name)
				fmt.Printf("Service deleted: %s/%s\n", svc.Namespace, svc.Name)
			},
		},
	)

	// Add event handlers to update mesh graph on Pod add/update/delete
	podInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					fmt.Println("Pod add: type assertion failed")
					return
				}
				svcKey := pod.Labels["app"]
				if svcKey != "" {
					fmt.Printf("Pod added: %s/%s (service: %s)\n", pod.Namespace, pod.Name, svcKey)
				} else {
					fmt.Printf("Pod added: %s/%s\n", pod.Namespace, pod.Name)
				}
				// TODO: Optionally associate pod with service in mesh graph
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod, ok := newObj.(*corev1.Pod)
				if !ok {
					fmt.Println("Pod update: type assertion failed")
					return
				}
				svcKey := pod.Labels["app"]
				if svcKey != "" {
					fmt.Printf("Pod updated: %s/%s (service: %s)\n", pod.Namespace, pod.Name, svcKey)
				} else {
					fmt.Printf("Pod updated: %s/%s\n", pod.Namespace, pod.Name)
				}
				// TODO: Optionally update pod association in mesh graph
			},
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					fmt.Println("Pod delete: type assertion failed")
					return
				}
				svcKey := pod.Labels["app"]
				if svcKey != "" {
					fmt.Printf("Pod deleted: %s/%s (service: %s)\n", pod.Namespace, pod.Name, svcKey)
				} else {
					fmt.Printf("Pod deleted: %s/%s\n", pod.Namespace, pod.Name)
				}
				// TODO: Optionally remove pod association from mesh graph
			},
		},
	)

	// TODO: Add informers for HTTPRoute, GRPCRoute, AuthorizationPolicy (requires CRD client-go codegen or dynamic client)

	// Subscribe to mesh:delta for policy reconciliation
	go func() {
		err := redis.SubscribeMeshDelta(context.Background(), func(msg []byte) {
			var patch graph.MeshGraph
			if err := json.Unmarshal(msg, &patch); err != nil {
				fmt.Printf("Collector: failed to unmarshal mesh delta: %v\n", err)
				return
			}
			// For demo: replace mesh with patch (real impl would merge/patch)
			mesh.AuthPolicies = patch.AuthPolicies
			fmt.Println("Collector: reconciled AuthPolicies from mesh delta")
			// TODO: Apply AuthPolicies to Kubernetes (create/update AuthorizationPolicy CRs)
			go func() {
				for {
					for key, policy := range mesh.AuthPolicies {
						fmt.Printf("Reconciling AuthorizationPolicy: %s\n", key)
						// Parse namespace and name from key
						var ns, name string
						parts := []rune(key)
						for i, c := range parts {
							if c == '/' {
								ns = string(parts[:i])
								name = string(parts[i+1:])
								break
							}
						}
						if ns == "" || name == "" {
							fmt.Printf("Invalid policy key: %s\n", key)
							continue
						}
						gvr := schema.GroupVersionResource{
							Group:    "policy.linkerd.io",
							Version:  "v1alpha1",
							Resource: "authorizationpolicies",
						}
						obj := &unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "policy.linkerd.io/v1alpha1",
								"kind":       "AuthorizationPolicy",
								"metadata": map[string]interface{}{
									"name":      name,
									"namespace": ns,
								},
								"spec": policy.Spec,
							},
						}
						// Try to create or update the AuthorizationPolicy
						_, err := dynClient.Resource(gvr).Namespace(ns).Get(context.Background(), name, metav1.GetOptions{})
						if err != nil {
							_, err = dynClient.Resource(gvr).Namespace(ns).Create(context.Background(), obj, metav1.CreateOptions{})
							if err != nil {
								fmt.Printf("Failed to create AuthorizationPolicy %s/%s: %v\n", ns, name, err)
							} else {
								fmt.Printf("Created AuthorizationPolicy %s/%s\n", ns, name)
							}
						} else {
							_, err = dynClient.Resource(gvr).Namespace(ns).Update(context.Background(), obj, metav1.UpdateOptions{})
							if err != nil {
								fmt.Printf("Failed to update AuthorizationPolicy %s/%s: %v\n", ns, name, err)
							} else {
								fmt.Printf("Updated AuthorizationPolicy %s/%s\n", ns, name)
							}
						}
					}
					time.Sleep(30 * time.Second)
				}
			}()
		})
		if err != nil {
			fmt.Printf("Collector: error subscribing to mesh:delta: %v\n", err)
		}
	}()

	// Initialize Prometheus client and poll metrics
	promClient, err := api.NewClient(api.Config{Address: cfg.PrometheusURL})
	if err != nil {
		fmt.Printf("Failed to create Prometheus client: %v\n", err)
		os.Exit(1)
	}
	v1api := promv1.NewAPI(promClient)
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			// Updated query for Linkerd 2.18+ proxy metrics
			query := `sum by(dst_namespace, dst_deployment, dst_service)(rate(request_total{direction="outbound"}[30s]))`
			result, warnings, err := v1api.Query(ctx, query, time.Now())
			if err != nil {
				fmt.Printf("Prometheus query error: %v\n", err)
			} else {
				if len(warnings) > 0 {
					fmt.Printf("Prometheus warnings: %v\n", warnings)
				}
				vector, ok := result.(model.Vector)
				if ok {
					var edges []graph.Edge
					for _, sample := range vector {
						dstNamespace := string(sample.Metric["dst_namespace"])
						dstDeployment := string(sample.Metric["dst_deployment"])
						dstService := string(sample.Metric["dst_service"])
						rps := float64(sample.Value)
						// Compose a unique identifier for the destination
						dst := dstNamespace + "/" + dstDeployment
						if dstService != "" {
							dst += "/" + dstService
						}
						edges = append(edges, graph.Edge{
							Src: "", // src is not available in this metric; can be extended if needed
							Dst: dst,
							RPS: rps,
							TLS: false, // TLS info not available in this metric; can be extended if needed
						})
					}
					mesh.Edges = edges
					fmt.Printf("Updated mesh.Edges with %d edges\n", len(edges))
				} else {
					fmt.Printf("Prometheus result: %v\n", result)
				}
			}
			time.Sleep(15 * time.Second)
		}
	}()

	// Periodically snapshot mesh graph to Redis
	go func() {
		for {
			snapshot, err := json.Marshal(mesh)
			if err != nil {
				fmt.Printf("Failed to marshal mesh graph: %v\n", err)
			} else {
				err := redis.SetMeshSnapshot(context.Background(), snapshot, 10*time.Minute)
				if err != nil {
					fmt.Printf("Failed to set mesh snapshot in Redis: %v\n", err)
				} else {
					fmt.Println("Published mesh snapshot to Redis")
				}
			}
			time.Sleep(30 * time.Second)
		}
	}()

	// Wait for signal (context cancellation)
	<-ctx.Done()
	fmt.Println("Shutting down MCP Collector...")
}
