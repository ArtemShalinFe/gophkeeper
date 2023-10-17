package config

import (
	"reflect"
	"testing"
)

const testString = "test"

func setTestEnvVars(t *testing.T) {
	t.Helper()

	t.Setenv("ADDRESS", testString)
	t.Setenv("CRYPTO_KEY", testString)
	t.Setenv("CERTIFICATE", testString)
}

func TestParseServerConfig(t *testing.T) {
	setTestEnvVars(t)

	tests := []struct {
		want    any
		cfg     *ServerCfg
		name    string
		wantErr bool
	}{
		{
			name: "check reading env",
			cfg:  NewConfig(),
			want: &ServerCfg{
				Addr:             testString,
				CertFilePath:     testString,
				PrivateCryptoKey: testString,
			},
			wantErr: false,
		},
		{
			name:    "check struct reading",
			want:    1,
			cfg:     nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := ReadEnvConfig(tt.cfg); err != nil {
				if !tt.wantErr {
					t.Errorf("ParseServerConfig() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.cfg, tt.want) {
				t.Errorf("ParseServerConfig() = %v, want %v", tt.cfg, tt.want)
			}
		})
	}
}
