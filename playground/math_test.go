package main

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(4, 5)
	if result != 20 {
		t.Errorf("expected 20, got %d", result)
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(50, 20)
	if result != 30 {
		t.Errorf("expected 30, got %d", result)
	}
}
