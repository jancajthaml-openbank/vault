package persistence

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

type storageMock struct {
	localfs.Storage
	files                     []string
	circuitBreakListDirectory bool
}

func (storage *storageMock) Chmod(absPath string, mod os.FileMode) error {
	return nil
}

func (storage *storageMock) ListDirectory(path string, acs bool) ([]string, error) {
	if storage.circuitBreakListDirectory {
		return nil, fmt.Errorf("list directory circuit break")
	}
	filtered := make([]string, len(storage.files))
	var ln int
	for _, file := range storage.files {
		if !strings.HasPrefix(file, path) {
			continue
		}
		parts := strings.Split(file, "/")
		filtered[ln] = parts[len(parts)-1]
		ln++
	}
	var result = make([]string, ln)
	copy(result, filtered[:ln])
	return result, nil
}

func (storage *storageMock) CountFiles(file string) (int, error) {
	return 0, nil
}

func (storage *storageMock) Exists(path string) (bool, error) {
	for _, file := range storage.files {
		if strings.HasPrefix(file, path) {
			return true, nil
		}
	}
	return false, nil
}

func (storage *storageMock) TouchFile(file string) error {
	storage.files = append(storage.files, file)
	return nil
}

func (storage *storageMock) ReadFileFully(file string) ([]byte, error) {
	return nil, nil
}

func (storage *storageMock) WriteFileExclusive(file string, data []byte) error {
	return nil
}

func (storage *storageMock) WriteFile(file string, data []byte) error {
	return nil
}

func (storage *storageMock) DeleteFile(file string) error {
	return nil
}

func (storage *storageMock) UpdateFile(file string, data []byte) error {
	return nil
}

func (storage *storageMock) AppendFile(file string, data []byte) error {
	return nil
}

func (storage storageMock) LastModification(string) (time.Time, error) {
	return time.Now(), nil
}

func TestLoadAccounts(t *testing.T) {
	t.Log("happy path")
	{
		storage := new(storageMock)
		storage.TouchFile("t_tenant/account/a")
		storage.TouchFile("/tmp/var")
		storage.TouchFile("t_tenant/account/b")
		storage.TouchFile("t_tenant/account/c")
		storage.TouchFile("/dev/null")
		accounts, err := LoadAccounts(storage, "tenant")
		require.Nil(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, accounts)
	}

	t.Log("misconfigured storage")
	{
		storage := new(storageMock)
		accounts, err := LoadAccounts(storage, "tenant")
		require.Nil(t, err)
		assert.Equal(t, []string{}, accounts)
	}

	t.Log("corrupted node")
	{
		storage := new(storageMock)
		storage.TouchFile("t_tenant/account/a")
		storage.circuitBreakListDirectory = true
		_, err := LoadAccounts(storage, "tenant")
		require.NotNil(t, err)
	}
}
