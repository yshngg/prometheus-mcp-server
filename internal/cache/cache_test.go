package cache

import (
	"testing"
	"time"
)

func TestCacheGetSet(t *testing.T) {
	c := New(time.Minute)
	c.Set("key", "value")
	v, ok := c.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if v.(string) != "value" {
		t.Fatalf("expected 'value', got %v", v)
	}
}

func TestCacheExpiry(t *testing.T) {
	c := New(10 * time.Millisecond)
	c.Set("key", "value")
	time.Sleep(20 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected key to be expired")
	}
}

func TestCacheMissing(t *testing.T) {
	c := New(time.Minute)
	_, ok := c.Get("nokey")
	if ok {
		t.Fatal("expected key to be missing")
	}
}
