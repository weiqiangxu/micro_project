package common

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGenerateJwtToken(t *testing.T) {
	type args struct {
		userId     uint64
		jwtSignKey string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "jwt create",
			args:    args{},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateJwtToken(tt.args.userId, tt.args.jwtSignKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJwtToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%#v", got)
		})
	}
}

func TestGetUid(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name:    "test get uid from jwt auth",
			args:    args{},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUid(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUid() got = %v, want %v", got, tt.want)
			}
		})
	}
}
