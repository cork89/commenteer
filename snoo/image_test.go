package snoo

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetImgProxy(t *testing.T) {
	t.Setenv("IMGPROXY_URL", "example.com")
	t.Setenv("IMGPROXY_SALT", "salt")
	t.Setenv("IMGPROXY_KEY", "key")

	want := "http://example.com/A0d0zvq27NeMDFzrS5pw5mAEkCpEljha_eQjnsyr-6E/resize:fit:1024:0:1/padding:0:0/wm:1:soea:0:0:0.3/background:255:255:255/plain/example.jpg"

	url, err := GetImgProxyUrl("example.jpg")

	if !cmp.Equal(url, want) || err != nil {
		t.Fatalf(`GetImgProxyUrl("example.jpg") = %s, want match for %s`, url, want)
	}
}
