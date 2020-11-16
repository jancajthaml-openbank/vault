package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountsPath(t *testing.T) {
	tenant := "tenant_1"

	path := AccountsPath(tenant)
	expected := fmt.Sprintf("t_%s/account", tenant)

	assert.Equal(t, expected, path)
}

func BenchmarkAccountsPath(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		AccountsPath("X")
	}
}
