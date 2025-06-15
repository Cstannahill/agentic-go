package vectorstore

import (
	"net/http"
	"testing"
)

func TestNewQdrantStoreWithHTTPClient(t *testing.T) {
	client := &http.Client{}
	qs := NewQdrantStore("http://localhost:6333", "c1", WithHTTPClient(client))
	if qs.Client != client {
		t.Fatalf("custom client not used")
	}
}
