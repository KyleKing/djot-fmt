package iohelper_test

import (
	"testing"

	"github.com/KyleKing/djot-fmt/internal/iohelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *iohelper.Options
		wantErr bool
	}{
		{
			name: "no args (stdin)",
			args: []string{},
			want: &iohelper.Options{
				SlwMarkers: ".!?",
				SlwWrap:    88,
				SlwMinLine: 40,
			},
		},
		{
			name: "input file only",
			args: []string{"input.djot"},
			want: &iohelper.Options{
				InputFile:  "input.djot",
				SlwMarkers: ".!?",
				SlwWrap:    88,
				SlwMinLine: 40,
			},
		},
		{
			name: "write flag",
			args: []string{"-w", "file.djot"},
			want: &iohelper.Options{
				Write:      true,
				InputFile:  "file.djot",
				SlwMarkers: ".!?",
				SlwWrap:    88,
				SlwMinLine: 40,
			},
		},
		{
			name: "check flag",
			args: []string{"-c", "file.djot"},
			want: &iohelper.Options{
				Check:      true,
				InputFile:  "file.djot",
				SlwMarkers: ".!?",
				SlwWrap:    88,
				SlwMinLine: 40,
			},
		},
		{
			name: "output flag",
			args: []string{"-o", "out.djot", "in.djot"},
			want: &iohelper.Options{
				OutputFile: "out.djot",
				InputFile:  "in.djot",
				SlwMarkers: ".!?",
				SlwWrap:    88,
				SlwMinLine: 40,
			},
		},
		{
			name: "no wrap sentences",
			args: []string{"--no-wrap-sentences", "file.djot"},
			want: &iohelper.Options{
				InputFile:       "file.djot",
				NoWrapSentences: true,
				SlwMarkers:      ".!?",
				SlwWrap:         88,
				SlwMinLine:      40,
			},
		},
		{
			name: "custom slw markers",
			args: []string{"--slw-markers", ".!?;", "file.djot"},
			want: &iohelper.Options{
				InputFile:  "file.djot",
				SlwMarkers: ".!?;",
				SlwWrap:    88,
				SlwMinLine: 40,
			},
		},
		{
			name: "custom slw wrap",
			args: []string{"--slw-wrap", "100", "file.djot"},
			want: &iohelper.Options{
				InputFile:  "file.djot",
				SlwMarkers: ".!?",
				SlwWrap:    100,
				SlwMinLine: 40,
			},
		},
		{
			name: "custom slw min line",
			args: []string{"--slw-min-line", "0", "file.djot"},
			want: &iohelper.Options{
				InputFile:  "file.djot",
				SlwMarkers: ".!?",
				SlwWrap:    88,
				SlwMinLine: 0,
			},
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
			got, err := iohelper.ParseArgs(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
