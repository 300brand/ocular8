package etcd

import (
	"testing"
)

func TestGetDefault(t *testing.T) {
	c := New("http://127.0.0.1:4001")

	// Ensure clean before and after testing
	c.Delete("/test", true)
	defer c.Delete("/test", true)

	exp := "value"
	value, err := c.GetDefault("/test/get", exp, "Test item")
	if err != nil {
		t.Fatalf("Get: %s", err)
	}
	if value != exp {
		t.Errorf("Did not get expected (%s) value: %s", exp, value)
	}

	value, err = c.GetDefault("/test/get", exp+"2", "")
	if err != nil {
		t.Fatalf("Get 2: %s", err)
	}
	if value != exp {
		t.Errorf("Did not get expected (%s) value: %s", exp, value)
	}
}

func TestGetAll(t *testing.T) {
	c := New("http://127.0.0.1:4001")

	// Ensure clean before and after testing
	c.Delete("/test", true)
	defer c.Delete("/test", true)

	configs := []*Item{
		&Item{
			Key:     "/test/mongo/dsn",
			Default: "mongodb://localhost:27017/ocular8",
			Desc:    "Connection string to MongoDB",
		},
		&Item{
			Key:     "/test/nsq/http",
			Default: "http://localhost:4151",
			Desc:    "NSQd HTTP address",
		},
		&Item{
			Key:     "/test/enqueue-feeds/limit",
			Default: "10",
			Desc:    "Max number of feeds to enqueue per batch",
		},
		&Item{
			Key:     "/test/enqueue-feeds/threshold",
			Default: "100",
			Desc:    "Entry threshold to avoid pushing more feeds into download queue. Applies to both feed and entry downloads.",
		},
	}
	if err := c.GetAll(configs); err != nil {
		t.Fatalf("GetAll: %s", err)
	}
	for _, item := range configs {
		if item.Value != item.Default {
			t.Errorf("Exp: %q Got: %q", item.Default, item.Value)
		}
	}
}
