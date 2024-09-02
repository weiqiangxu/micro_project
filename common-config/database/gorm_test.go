package database

import (
	"reflect"
	"testing"

	"github.com/weiqiangxu/common-config/format"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestInitGormV2(t *testing.T) {
	type args struct {
		cfg        *format.MysqlConfig
		gormLogger logger.Interface
	}
	tests := []struct {
		name    string
		args    args
		want    *gorm.DB
		wantErr bool
	}{
		{
			name:    "test g orm",
			args:    args{},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitGormV2(tt.args.cfg, tt.args.gormLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitGormV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitGormV2() got = %v, want %v", got, tt.want)
			}
		})
	}
}
