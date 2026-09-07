package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"jerremiah.dev/llmservice-cli/internal/llmservice"
)

func TestOrDefault(t *testing.T) {
	cases := []struct {
		value, fieldKey, want string
	}{
		{"", "image", "(not set - CRD default: vllm/vllm-openai:latest)"},
		{"custom-image:v1", "image", "custom-image:v1"},
		{"", "cpuRequest", "(not set - CRD default: 100m)"},
		{"", "some-field-with-no-default", "(not set)"},
	}
	for _, c := range cases {
		got := orDefault(c.value, c.fieldKey)
		if got != c.want {
			t.Errorf("orDefault(%q, %q) = %q, want %q", c.value, c.fieldKey, got, c.want)
		}
	}
}

func TestInt32PtrOrNotSet(t *testing.T) {
	if got := int32PtrOrNotSet(nil); got != "(not set)" {
		t.Errorf("int32PtrOrNotSet(nil) = %q, want %q", got, "(not set)")
	}
	v := int32(5)
	if got := int32PtrOrNotSet(&v); got != "5" {
		t.Errorf("int32PtrOrNotSet(&5) = %q, want %q", got, "5")
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns
// whatever was written to it. printStatus uses plain fmt.Printf (writes
// to os.Stdout directly) rather than taking an io.Writer, so this is the
// most direct way to test its output without refactoring the function
// signature.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = original

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	return string(out)
}

func TestPrintStatus_ShowsCRDDefaultsForUnsetFields(t *testing.T) {
	spec := llmservice.Spec{Model: "meta-llama/Llama-3-8b"} // everything else unset
	status := llmservice.Status{}

	output := captureStdout(t, func() { printStatus("my-model", spec, status) })

	if !strings.Contains(output, "CRD default: vllm/vllm-openai:latest") {
		t.Errorf("expected image default in output, got:\n%s", output)
	}
	if !strings.Contains(output, "CRD default: 100m") {
		t.Errorf("expected cpu request default in output, got:\n%s", output)
	}
	if !strings.Contains(output, "replicas:       (not set)") {
		t.Errorf("expected replicas to show (not set), got:\n%s", output)
	}
}

func TestPrintStatus_ShowsExplicitValuesWhenSet(t *testing.T) {
	spec := llmservice.Spec{
		Model:    "meta-llama/Llama-3-8b",
		Image:    "vllm/vllm-openai:v0.5.0",
		GPUCount: int32Ptr(2),
	}
	status := llmservice.Status{Phase: "Running"}

	output := captureStdout(t, func() { printStatus("my-model", spec, status) })

	if !strings.Contains(output, "vllm/vllm-openai:v0.5.0") {
		t.Errorf("expected explicit image value in output, got:\n%s", output)
	}
	if strings.Contains(output, "CRD default: vllm/vllm-openai:latest") {
		t.Errorf("should NOT show the default once an explicit image is set, got:\n%s", output)
	}
	if !strings.Contains(output, "gpu count:      2") {
		t.Errorf("expected gpu count 2 in output, got:\n%s", output)
	}
	if !strings.Contains(output, "phase:          Running") {
		t.Errorf("expected phase Running in output, got:\n%s", output)
	}
}
