package main

import (
	"context"
	"testing"

	"github.com/jellydator/ttlcache/v3"
)

func BenchmarkGetToken(b *testing.B) {
	// Context for testing
	ctx := context.Background()

	// Reset cache
	bboneTokenCache = ttlcache.New[string, *GetTokenResDto]()

	// Benchmark loop
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetToken(ctx)
		if err != nil {
			b.Fatalf("GetToken failed: %v", err)
		}
	}
}
