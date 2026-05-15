package tui

import "testing"

func TestCommandForFileManager(t *testing.T) {
	tests := []struct {
		name string
		goos string
		want string
	}{
		{name: "windows", goos: "windows", want: "explorer"},
		{name: "macos", goos: "darwin", want: "open"},
		{name: "linux", goos: "linux", want: "xdg-open"},
		{name: "other unix", goos: "freebsd", want: "xdg-open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := commandForFileManager(tt.goos, "/tmp/repo")
			if cmd.Args[0] != tt.want {
				t.Fatalf("command = %q, want %q", cmd.Args[0], tt.want)
			}
			if len(cmd.Args) != 2 || cmd.Args[1] != "/tmp/repo" {
				t.Fatalf("command args = %#v, want path argument", cmd.Args)
			}
		})
	}
}
