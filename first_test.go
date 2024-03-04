package main

import "testing"

func TestFirst(t *testing.T) {
	x := 5
	y := 5
	if x != y {
		t.Errorf("Expected variables to be equal!")
	}
}
