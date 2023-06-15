package main

import (
	"testing"

	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/require"
)

func TestLoggerConfig_Validate(t *testing.T) {
	type fields struct {
		LogLevel  string
		LogFormat string
	}
	tests := []struct {
		name     string
		fields   fields
		wantFunc func(t *testing.T, err error)
	}{
		{
			name: "bad log level",
			fields: fields{
				LogLevel:  "notALogLevel",
				LogFormat: "logfmt",
			},
			wantFunc: func(t *testing.T, err error) {
				require.ErrorIs(t, err, level.ErrInvalidLevelString, "log level should be invalid")
			},
		},
		{
			name: "bad log format",
			fields: fields{
				LogLevel:  "debug",
				LogFormat: "notALogFormat",
			},
			wantFunc: func(t *testing.T, err error) {
				require.ErrorIs(t, err, errInvalidLogFormatFlag, "log format flag should be invalid")
			},
		},
		{
			name: "good config",
			fields: fields{
				LogLevel:  "debug",
				LogFormat: "json",
			},
			wantFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lc := &LoggerConfig{
				LogLevel:  tt.fields.LogLevel,
				LogFormat: tt.fields.LogFormat,
			}
			err := lc.Validate()
			tt.wantFunc(t, err)
		})
	}
}
