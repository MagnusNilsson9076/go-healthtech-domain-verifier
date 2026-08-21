package main

import (
	"context"
	"fmt"
	"os"

	domainverifier "healthtech-domain-verifier"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/verify-domain example.com")
		os.Exit(2)
	}
	client, err := domainverifier.NewClient(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	verification, err := client.VerifyDomain(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("domain=%s verification.status=%s\n", os.Args[1], verification.Status)
}
