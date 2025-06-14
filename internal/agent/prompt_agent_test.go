package agent

import "testing"

func TestPromptAgent(t *testing.T) {
	a := NewPromptAgent()
	task := Task{ID: "1", Input: map[string]interface{}{
		"template": "Hello {{.User}}",
		"context":  map[string]interface{}{"User": "World"},
	}}
	res := a.Execute(nil, task)
	if !res.Successful || res.Output.(map[string]interface{})["prompt"] != "Hello World" {
		t.Fatalf("unexpected result: %+v", res)
	}
}
