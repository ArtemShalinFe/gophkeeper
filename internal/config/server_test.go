package config

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

var testString = uuid.NewString()

func TestReadEnvServerCfg(t *testing.T) {
	t.Setenv("GKS_ADDRESS", testString)
	t.Setenv("CRYPTO_KEY", testString)
	t.Setenv("CERTIFICATE", testString)
	t.Setenv("DATABASE_DSN", testString)

	tests := []struct {
		want    any
		cfg     *ServerCfg
		name    string
		wantErr bool
	}{
		{
			name: "check reading env",
			cfg:  NewServerCfg(),
			want: &ServerCfg{
				Addr:             testString,
				CertFilePath:     testString,
				PrivateCryptoKey: testString,
				DSN:              testString,
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
			if err := ReadEnvServerCfg(tt.cfg); err != nil {
				if !tt.wantErr {
					t.Errorf("ReadEnvServerCfg() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.cfg, tt.want) {
				t.Errorf("ReadEnvServerCfg() = %v, want %v", tt.cfg, tt.want)
			}
		})
	}
}
