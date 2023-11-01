package config

import (
	"reflect"
	"testing"
)

func TestReadEnvClientCfg(t *testing.T) {
	t.Setenv("GKS_ADDRESS", testString)
	t.Setenv("AGENT_KEY", testString)
	t.Setenv("CERTIFICATE", testString)

	tests := []struct {
		want    any
		cfg     *ClientCfg
		name    string
		keyword string
		wantErr bool
	}{
		{
			name: "check reading env",
			cfg:  NewClientCfg(),
			want: &ClientCfg{
				GKeeper:      testString,
				CertFilePath: testString,
				KeyFilePath:  testString,
				Keyword:      []byte(testString),
			},
			keyword: testString,
			wantErr: false,
		},
		{
			name: "check reading env",
			cfg:  NewClientCfg(),
			want: &ClientCfg{
				GKeeper:      testString,
				CertFilePath: testString,
				KeyFilePath:  testString,
				Keyword:      []byte(defaultKeyword),
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
			if tt.keyword == "" {
				t.Setenv("KEYWORD", defaultKeyword)
			}
			t.Setenv("KEYWORD", tt.keyword)

			if err := ReadEnvClientCfg(tt.cfg); err != nil {
				if !tt.wantErr {
					t.Errorf("ReadEnvClientCfg() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.cfg, tt.want) {
				t.Errorf("ReadEnvClientCfg() = %v, want %v", tt.cfg, tt.want)
			}
		})
	}
}
