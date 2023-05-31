package design

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"
)

func TestNewTinyDB(t *testing.T) {
	tests := []struct {
		name string
		want TinyDB
	}{
		{"simply new", &redis{data: map[string]any{}, RWMutex: sync.RWMutex{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTinyDB(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTinyDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_redis_Del(t *testing.T) {
	tests := []struct {
		name        string
		executeFunc func() TinyDB
		key         string
		wantErr     error
	}{
		{"simply del",
			func() TinyDB {
				r := NewTinyDB()
				r.Set("key", "val")
				return nil
			},
			"key", nil},
	}

	r := NewTinyDB()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.Del(tt.key)
			assert.Equal(t, tt.wantErr, r.Del("key"))
		})
	}
}

func Test_redis_Get(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		val     interface{}
		wantErr error
	}{
		{"simply get", "key", "val", nil},
		{"get failed", "key2", "val2", redisKeyNotFound},
	}
	r := NewTinyDB()
	r.Set("key", "val")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Get(tt.key)
			assert.Equal(t, tt.wantErr, err)

			if err == nil && !reflect.DeepEqual(got, tt.val) {
				t.Errorf("Get() got = %v, want %v", got, tt.val)
			}
		})
	}
}

func Test_redis_Set(t *testing.T) {

	tests := []struct {
		name  string
		key   string
		value any
	}{
		{"simply set", "key", "val"},
	}
	r := NewTinyDB()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, r.Set(tt.key, tt.value))
		})
	}
}
