package v1

import (
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	type args struct {
		ttl time.Duration
	}
	tests := []struct {
		name string
		args args
		want *Cache
	}{
		{name: "testnew_", args: args{ttl: time.Duration(time.Second)}, want: New(time.Duration(time.Second))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.ttl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_Set(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "TestCache_Set_1",
			args: args{
				key:   "test1",
				value: 123,
			},
			want: 123,
		},
		{
			name: "TestCache_Set_2",
			args: args{
				key:   "test2",
				value: `{"name":"jeffotoni"}`,
			},
			want: `{"name":"jeffotoni"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(200 * time.Millisecond)
			c.Set(tt.args.key, tt.args.value, NoExpiration)
			got, exist := c.Get(tt.args.key)
			if exist {
				switch got.(type) {
				case int:
					vgot := got.(int)
					if vgot != tt.want {
						t.Errorf("Cache.Get() = %v, want %v", vgot, tt.want)
					}
				case string:
					vgot := got.(string)
					if vgot != tt.want {
						t.Errorf("Cache.Get() = %v, want %v", vgot, tt.want)
					}
				}
			}
			time.Sleep(300 * time.Millisecond)
			_, exist = c.Get(tt.args.key)
			if !exist {
				t.Errorf("Cache item should have been expired and not exist")
			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	type args struct {
		key   string
		value interface{}
		ttl   time.Duration
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		found bool
	}{
		{
			name:  "Get existing item jeffotoni",
			args:  args{key: "key1", value: 123, ttl: DefaultExpiration},
			want:  123,
			found: true,
		},
		{
			name:  "Get non-existent item jeffotoni",
			args:  args{key: "nonExistent", value: nil, ttl: DefaultExpiration},
			want:  nil,
			found: false,
		},
		{
			name:  "Get expired item jeffotoni",
			args:  args{key: "expiredKey", value: "expired", ttl: 1 * time.Second},
			want:  nil,
			found: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(DefaultExpiration)
			if tt.args.value != nil {
				c.Set(tt.args.key, tt.args.value, tt.args.ttl)
			}
			if tt.args.ttl > 0 {
				time.Sleep(2 * time.Second)
			}
			got, found := c.Get(tt.args.key)
			if got != tt.want || found != tt.found {
				t.Errorf("Cache.Get() = %v, %v, want %v, %v", got, found, tt.want, tt.found)
			}
		})
	}
}

func TestCache_Delete(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name  string
		set   []args
		del   args
		found bool
	}{
		{
			name:  "Delete existing key",
			set:   []args{{key: "key1"}},
			del:   args{key: "key1"},
			found: false,
		},
		{
			name:  "Delete non-existent key",
			set:   []args{{key: "key2"}},
			del:   args{key: "key999"},
			found: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(DefaultExpiration)
			for _, s := range tt.set {
				c.Set(s.key, "value", DefaultExpiration)
			}
			c.Delete(tt.del.key)
			_, found := c.Get(tt.del.key)
			if found != tt.found {
				t.Errorf("Cache.Get() after Delete = %v, want %v", found, tt.found)
			}
		})
	}
}
func TestCache_clean(t *testing.T) {
	tests := []struct {
		name string
		set  []struct {
			key   string
			value interface{}
			ttl   time.Duration
		}
		waitTime time.Duration
		wantKeys []string
	}{
		{
			name: "Clean expired keys",
			set: []struct {
				key   string
				value interface{}
				ttl   time.Duration
			}{
				{key: "key1", value: "value1", ttl: 400 * time.Millisecond},
				{key: "key2", value: "value2", ttl: 500 * time.Millisecond},
			},
			waitTime: 3 * time.Millisecond,
			wantKeys: []string{}, // Todos os itens devem expirar
		},
		{
			name: "Keep valid keys after clean",
			set: []struct {
				key   string
				value interface{}
				ttl   time.Duration
			}{
				{key: "validKey", value: "value3", ttl: 5 * time.Millisecond},
				{key: "expiredKey", value: "value4", ttl: 1 * time.Millisecond},
			},
			waitTime: 3 * time.Millisecond,
			wantKeys: []string{"validKey"}, // Apenas "validKey" deve permanecer
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(DefaultExpiration)
			for _, item := range tt.set {
				c.Set(item.key, item.value, item.ttl)
			}
			time.Sleep(tt.waitTime)
			c.clean()
			for _, key := range tt.wantKeys {
				_, found := c.Get(key)
				if !found {
					t.Errorf("Cache.clean() missing expected key: %v", key)
				}
			}
		})
	}
}
