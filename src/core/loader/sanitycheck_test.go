package loader

import "testing"

func TestIsGoRunBuildDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "temporary go run build", path: "/tmp/go-build123456/b001/exe", want: true},
		{name: "cached go run build", path: "/home/vscode/.cache/go-build/ab/abcdef-d", want: true},
		{name: "go build cache root", path: "/home/vscode/.cache/go-build", want: true},
		{name: "installed binary", path: "/opt/ssui", want: false},
		{name: "similarly named directory", path: "/opt/go-builder/ssui", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGoRunBuildDir(tt.path); got != tt.want {
				t.Fatalf("isGoRunBuildDir(%q) = %t, want %t", tt.path, got, tt.want)
			}
		})
	}
}
