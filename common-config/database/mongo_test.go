package database

import (
	"reflect"
	"testing"

	"github.com/weiqiangxu/common-config/format"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestInitMongo(t *testing.T) {
	type args struct {
		config *format.MongoConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *mongo.Database
		wantErr bool
	}{
		{
			name:    "test mongodb driver",
			args:    args{},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitMongo(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitMongo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitMongo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
