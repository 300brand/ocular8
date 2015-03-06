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

	c.Set("/test/k1", "v1", 0)
	c.Set("/test/k2", "v2", 0)

	var v1, v2 string
	kv := map[string]*string{
		"/test/k1": &v1,
		"/test/k2": &v2,
	}
	if err := c.GetAll(kv); err != nil {
		t.Fatalf("GetAll: %s", err)
	}
	if v1 != "v1" {
		t.Errorf("v1 = %q", v1)
	}
	if v2 != "v2" {
		t.Errorf("v2 = %q", v1)
	}
}
