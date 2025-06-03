// demo-app/main.go

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func runServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from service-b!")
	})
	addr := ":8080"
	log.Printf("Starting service-b (server) on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func runClient(target string) {
	for {
		resp, err := http.Get(fmt.Sprintf("http://%s:8080/", target))
		if err != nil {
			log.Printf("service-a: error calling service-b: %v", err)
		} else {
			log.Printf("service-a: called service-b, got status: %s", resp.Status)
			resp.Body.Close()
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	mode := flag.String("mode", "", "Mode: 'a' for client, 'b' for server")
	target := flag.String("target", "service-b", "Target service name for client mode")
	flag.Parse()

	if *mode == "" {
		// Allow env var override for Kubernetes
		*mode = os.Getenv("DEMO_MODE")
	}
	if *mode == "" {
		log.Fatal("Must specify --mode=a or --mode=b (or set DEMO_MODE env var)")
	}

	if *mode == "b" {
		runServer()
	} else if *mode == "a" {
		runClient(*target)
	} else {
		log.Fatalf("Unknown mode: %s", *mode)
	}
}
