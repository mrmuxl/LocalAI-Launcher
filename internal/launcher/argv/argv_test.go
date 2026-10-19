package argv

import (
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      Cfg
		wantErr   bool
		wantErrIs error
		// wantErrContains 列出错误信息中必须包含的片段
		wantErrContains []string
	}{
		{
			name: "empty args returns zero cfg",
			args: []string{},
			want: Cfg{},
		},
		{
			name: "model only",
			args: []string{"--model=qwen2.5-7b"},
			want: Cfg{Model: "qwen2.5-7b"},
		},
		{
			name: "device only",
			args: []string{"--device=gpu"},
			want: Cfg{Device: "gpu"},
		},
		{
			name: "host only",
			args: []string{"--host=0.0.0.0"},
			want: Cfg{Host: "0.0.0.0"},
		},
		{
			name: "port only",
			args: []string{"--port=8080"},
			want: Cfg{Port: 8080},
		},
		{
			name: "threads only",
			args: []string{"--threads=4"},
			want: Cfg{Threads: 4},
		},
		{
			name: "multiple flags combined",
			args: []string{
				"--model=qwen2.5-7b",
				"--device=cpu",
				"--host=127.0.0.1",
				"--port=9090",
				"--threads=8",
			},
			want: Cfg{Model: "qwen2.5-7b", Device: "cpu", Host: "127.0.0.1", Port: 9090, Threads: 8},
		},
		{
			name:            "unknown flag",
			args:            []string{"--bogus=1"},
			wantErr:         true,
			wantErrContains: []string{"--bogus"},
		},
		{
			name:            "positional argument is unknown flag",
			args:            []string{"just-a-value"},
			wantErr:         true,
			wantErrContains: []string{"just-a-value"},
		},
		{
			name:            "port non-numeric",
			args:            []string{"--port=abc"},
			wantErr:         true,
			wantErrContains: []string{"--port", "abc"},
		},
		{
			name:            "port out of range",
			args:            []string{"--port=70000"},
			wantErr:         true,
			wantErrContains: []string{"--port", "70000"},
		},
		{
			name:            "threads negative",
			args:            []string{"--threads=-1"},
			wantErr:         true,
			wantErrContains: []string{"--threads", "-1"},
		},
		{
			name:            "threads zero",
			args:            []string{"--threads=0"},
			wantErr:         true,
			wantErrContains: []string{"--threads", "0"},
		},
		{
			name:            "threads non-numeric",
			args:            []string{"--threads=xyz"},
			wantErr:         true,
			wantErrContains: []string{"--threads", "xyz"},
		},
		{
			name:      "short help flag",
			args:      []string{"-h"},
			wantErr:   true,
			wantErrIs: ErrHelp,
		},
		{
			name:      "long help flag",
			args:      []string{"--help"},
			wantErr:   true,
			wantErrIs: ErrHelp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%v) expected error, got nil", tt.args)
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("Parse(%v) expected ErrHelp, got %v", tt.args, err)
				}
				for _, s := range tt.wantErrContains {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("Parse(%v) error %q does not contain %q", tt.args, err.Error(), s)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse(%v) unexpected error: %v", tt.args, err)
			}
			if got != tt.want {
				t.Fatalf("Parse(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}
