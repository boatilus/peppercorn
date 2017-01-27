package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const ip = "192.168.0.1"
const ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36"

func TestCreateID(t *testing.T) {
	got := createID(ip + ua)
	assert.NotEmpty(t, got)
}

func BenchmarkCreateID(b *testing.B) {
	param := ip + ua

	for n := 0; n < b.N; n++ {
		createID(param)
	}
}
