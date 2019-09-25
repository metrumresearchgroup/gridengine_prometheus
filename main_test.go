package main

import (
	"os"
	"testing"
)

func TestWritePidFile(t *testing.T) {
	normal := "/tmp/meow.pid"
	invalid := "/foo/bar/baz/qux.pid"
	type args struct {
		pidFile *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Normal location",
			args: args{
				pidFile: &normal,
			},
			wantErr: false,
		},
		{
			name: "Invalid location",
			args: args{
				pidFile: &invalid,
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

func TestSetup(t *testing.T) {
	os.Unsetenv("TEST")
	setup()

	if isTest {
		t.Errorf("Test should not be set at this point")
	}

	random.Int()
	random.Float64()

	os.Setenv("TEST", "true")

	setup()

	if !isTest {
		t.Errorf("Test mode should be set now")
	}

	os.Setenv("LISTEN_PORT", "9014")
	setup()
}
