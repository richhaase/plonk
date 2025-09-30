package packages

import "testing"

func TestCondaParseSearchOutput(t *testing.T) {
	c := NewCondaManager()
	out := []byte(`{"numpy":[{"name":"numpy","version":"1.26.0"}],"pandas":[{"name":"pandas","version":"2.2.0"}]}`)
	res := c.parseSearchOutput(out)
	if len(res) != 2 {
		t.Fatalf("want 2, got %v", res)
	}
}

func TestCondaParseInfoOutput(t *testing.T) {
	c := NewCondaManager()
	out := []byte(`{"numpy":[{"name":"numpy","version":"1.26.0","summary":"NumPy","home":"https://numpy.org","depends":["python >=3.9"]}]}`)
	info := c.parseInfoOutput(out, "numpy")
	if info == nil || info.Version != "1.26.0" {
		t.Fatalf("bad info: %+v", info)
	}
}
