package iohelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Options
		wantErr bool
	}{
		{
			name: "no args (stdin)",
			args: []string{},
			want: &Options{},
		},
		{
			name: "input file only",
			args: []string{"input.djot"},
			want: &Options{InputFile: "input.djot"},
		},
		{
			name: "write flag",
			args: []string{"-w", "file.djot"},
			want: &Options{Write: true, InputFile: "file.djot"},
		},
		{
			name: "check flag",
			args: []string{"-c", "file.djot"},
			want: &Options{Check: true, InputFile: "file.djot"},
		},
		{
			name: "output flag",
			args: []string{"-o", "out.djot", "in.djot"},
			want: &Options{OutputFile: "out.djot", InputFile: "in.djot"},
		},
		{
			name:    "write without file",
			args:    []string{"-w"},
			wantErr: true,
		},
		{
			name:    "write and output",
			args:    []string{"-w", "-o", "out.djot", "in.djot"},
			wantErr: true,
		},
		{
			name:    "check and write",
			args:    []string{"-c", "-w", "file.djot"},
			wantErr: true,
		},
		{
			name:    "unknown flag",
			args:    []string{"-x"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
