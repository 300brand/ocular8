package etcd

import (
	"testing"
)

func TestGetDefault(t *testing.T) {
	c := New([]string{"http://127.0.0.1:4001"})

	// Ensure clean before and after testing
	c.Delete("/test", true)
	// defer c.Delete("/test", true)

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
