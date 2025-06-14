package tools

import "testing"

func TestBasicHashEmbed(t *testing.T) {
	v1 := BasicHashEmbed("hello world", 10)
	v2 := BasicHashEmbed("hello world", 10)
	if len(v1) != 10 || len(v2) != 10 {
		t.Fatalf("unexpected dimension")
	}
	for i := range v1 {
		if v1[i] != v2[i] {
			t.Fatalf("embedding not deterministic")
		}
	}
}
