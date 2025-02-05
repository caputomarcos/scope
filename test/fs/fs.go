package fs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/common/fs"
)

type mockInode struct{}

type dir struct {
	mockInode
	name    string
	entries map[string]Entry
	stat    syscall.Stat_t
}

// File is a mock file
type File struct {
	mockInode
	FName     string
	FContents string
	FStat     syscall.Stat_t
}

// Entry is an entry in the mock filesystem
type Entry interface {
	os.FileInfo
	fs.Interface
}

// Dir creates a new directory with the given entries.
func Dir(name string, entries ...Entry) Entry {
	result := dir{
		name:    name,
		entries: map[string]Entry{},
	}

	for _, entry := range entries {
		result.entries[entry.Name()] = entry
	}

	return result
}

func split(path string) (string, string) {
	if !strings.HasPrefix(path, "/") {
		panic(path)
	}

	comps := strings.SplitN(path, "/", 3)
	if len(comps) == 2 {
		return comps[1], "/"
	}

	return comps[1], "/" + comps[2]
}

func (mockInode) Size() int64        { return 0 }
func (mockInode) Mode() os.FileMode  { return 0 }
func (mockInode) ModTime() time.Time { return time.Now() }
func (mockInode) Sys() interface{}   { return nil }

func (p dir) Name() string { return p.name }
func (p dir) IsDir() bool  { return true }

func (p dir) ReadDir(path string) ([]os.FileInfo, error) {
	if path == "/" {
		result := []os.FileInfo{}
		for _, v := range p.entries {
			result = append(result, v)
		}
		return result, nil
	}

	head, tail := split(path)
	fs, ok := p.entries[head]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", path)
	}

	return fs.ReadDir(tail)
}

func (p dir) ReadDirNames(path string) ([]string, error) {
	if path == "/" {
		result := []string{}
		for _, v := range p.entries {
			result = append(result, v.Name())
		}
		return result, nil
	}

	head, tail := split(path)
	fs, ok := p.entries[head]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", path)
	}

	return fs.ReadDirNames(tail)
}

func (p dir) ReadFile(path string) ([]byte, error) {
	if path == "/" {
		return nil, fmt.Errorf("I'm a directory!")
	}

	head, tail := split(path)
	fs, ok := p.entries[head]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", path)
	}

	return fs.ReadFile(tail)
}

func (p dir) Lstat(path string, stat *syscall.Stat_t) error {
	if path == "/" {
		return nil
	}

	head, tail := split(path)
	fs, ok := p.entries[head]
	if !ok {
		return fmt.Errorf("Not found: %s", path)
	}

	return fs.Lstat(tail, stat)
}

func (p dir) Stat(path string, stat *syscall.Stat_t) error {
	if path == "/" {
		return nil
	}

	head, tail := split(path)
	fs, ok := p.entries[head]
	if !ok {
		return fmt.Errorf("Not found: %s", path)
	}

	return fs.Stat(tail, stat)
}

func (p dir) Open(path string) (io.ReadWriteCloser, error) {
	if path == "/" {
		return nil, fmt.Errorf("I'm a directory!")
	}

	head, tail := split(path)
	fs, ok := p.entries[head]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", path)
	}

	return fs.Open(tail)
}

// Name implements os.FileInfo
func (p File) Name() string { return p.FName }

// IsDir implements os.FileInfo
func (p File) IsDir() bool { return false }

// ReadDir implements FS
func (p File) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("I'm a file!")
}

// ReadDirNames implements FS
func (p File) ReadDirNames(path string) ([]string, error) {
	return nil, fmt.Errorf("I'm a file!")
}

// ReadFile implements FS
func (p File) ReadFile(path string) ([]byte, error) {
	if path != "/" {
		return nil, fmt.Errorf("I'm a file!")
	}
	return []byte(p.FContents), nil
}

// Lstat implements FS
func (p File) Lstat(path string, stat *syscall.Stat_t) error {
	if path != "/" {
		return fmt.Errorf("I'm a file!")
	}
	*stat = p.FStat
	return nil
}

// Stat implements FS
func (p File) Stat(path string, stat *syscall.Stat_t) error {
	if path != "/" {
		return fmt.Errorf("I'm a file!")
	}
	*stat = p.FStat
	return nil
}

// Open implements FS
func (p File) Open(path string) (io.ReadWriteCloser, error) {
	if path != "/" {
		return nil, fmt.Errorf("I'm a file!")
	}
	return struct {
		io.ReadWriter
		io.Closer
	}{
		bytes.NewBuffer([]byte(p.FContents)),
		ioutil.NopCloser(nil),
	}, nil
}
