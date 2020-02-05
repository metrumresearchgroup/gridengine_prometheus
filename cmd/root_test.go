package cmd

import (
	"testing"
)

func TestWritePidFile(t *testing.T) {
	normal := "/tmp/meow.pid"
	invalid := "/foo/bar/baz/qux.pid"
	type args struct {
		pidFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Normal location",
			args: args{
				pidFile: normal,
			},
			wantErr: false,
		},
		{
			name: "Invalid location",
			args: args{
				pidFile: invalid,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writePidFile(tt.args.pidFile); (err != nil) != tt.wantErr {
				t.Errorf("writePidFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}