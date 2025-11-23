package cache

import (
	"testing"
	"wb-l0/internal/model"
)

func TestCache(t *testing.T) {
	c := New()
	order := model.Order{OrderUID: "123", TrackNumber: "test"}

	c.Set(order)

	got, err := c.Get("123")
	if err != nil {
		t.Errorf("Expected to get order, got error: %v", err)
	}
	if got.TrackNumber != "test" {
		t.Errorf("Expected track 'test', got %s", got.TrackNumber)
	}

	_, err = c.Get("999")
	if err == nil {
		t.Error("Expected error for missing id, got nil")
	}
}
