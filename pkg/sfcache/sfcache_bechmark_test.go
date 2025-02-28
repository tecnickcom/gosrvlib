package sfcache

import (
	"context"
	"strconv"
	"testing"
	"time"
)

const testDomain = "example.com"

func BenchmarkLookup_cache_miss(b *testing.B) {
	lookupFn := func(_ context.Context, _ string) (any, error) {
		return []string{"192.0.2.1"}, nil
	}

	c := New(lookupFn, int(1<<63-1), 1*time.Second)

	b.ResetTimer()

	for i := range b.N {
		_, _ = c.Lookup(b.Context(), strconv.Itoa(i)+testDomain)
	}
}

func BenchmarkLookup_cache_hit(b *testing.B) {
	lookupFn := func(_ context.Context, _ string) (any, error) {
		return []string{"192.0.2.1"}, nil
	}

	size := 255

	c := New(lookupFn, size, 1*time.Minute)

	// fill the cache
	for i := 1; i <= size; i++ {
		_, _ = c.Lookup(b.Context(), strconv.Itoa(i)+testDomain)
	}

	var j int

	b.ResetTimer()

	for range b.N {
		j++
		if j > size {
			j = 0
		}

		_, _ = c.Lookup(b.Context(), strconv.Itoa(j)+testDomain)
	}
}
