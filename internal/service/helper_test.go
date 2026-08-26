package service

import (
	"fmt"

	"task264-geneclock/internal/store"
)

// openMemoryStore 打开 SQLite 内存数据库（file::memory: 模式）。
func openMemoryStore() (*store.Store, error) {
	st, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("open memory store: %w", err)
	}
	return st, nil
}
