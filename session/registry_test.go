package session

import (
	"testing"

	"gopkg.in/redis.v3"
)

func TestMemory(t *testing.T) {
	r := newRegistry("hoge", MemoryDialect, nil)
	defer r.dispose()

	err := r.set("fuga", []byte("piyopiyo"), 0)

	if err != nil {
		t.Errorf("expected nil, actual %v", err)
	}

	val, err := r.get("fuga")

	if err != nil {
		t.Errorf("expected nil, actual %v", err)
	}

	if string(val) != "piyopiyo" {
		t.Errorf("expected piyopiyo, actual %v", val)
	}

}

func TestRedis(t *testing.T) {
	r := newRegistry("exampleNamespace", RedisDialect, &redis.Options{Addr: "localhost:6379"})
	defer r.dispose()

	err := r.set("fuga", []byte("piyopiyo"), 0)

	if err != nil {
		t.Errorf("expected nil, actual %v", err)
	}

	val, err := r.get("fuga")

	if err != nil {
		t.Errorf("expected nil, actual %v", err)
	}

	if string(val) != "piyopiyo" {
		t.Errorf("expected piyopiyo, actual %v", val)
	}
}
