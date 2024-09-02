package redisapi

import (
	"reflect"
	"testing"

	"github.com/weiqiangxu/common-config/format"
)

func TestNewRedisApi(t *testing.T) {
	type args struct {
		redisConfig format.RedisConfig
	}
	tests := []struct {
		name string
		args args
		want *RedisApi
	}{
		{
			name: "",
			args: args{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRedisApi(tt.args.redisConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRedisApi() = %v, want %v", got, tt.want)
			}
		})
	}
}
