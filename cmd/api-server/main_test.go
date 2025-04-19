package main

import "testing"

func TestStringGen(t *testing.T) {
	result := StringGen()
	if result != "" {
		t.Errorf("aFx should return '', returned: '%s'", result)
	}
}
