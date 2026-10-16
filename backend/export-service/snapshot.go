package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

var (
	bucketOffers = []byte("offers")
	bucketMeta   = []byte("meta")
)

// SnapshotStore хранит инкрементальный слепок предсериализованных XML-фрагментов офферов
// в embedded ACID Key-Value базе bbolt на постоянном диске.
type SnapshotStore struct {
	db   *bbolt.DB
	path string
	mu   sync.RWMutex
}

// NewSnapshotStore открывает или создает локальную базу bbolt для хранения слепка каталога
func NewSnapshotStore(dbPath string) (*SnapshotStore, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("создать директорию для базы слепка: %w", err)
	}

	db, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("открыть базу слепка bbolt (%s): %w", dbPath, err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketOffers); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketMeta); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("инициализировать buckets bbolt: %w", err)
	}

	return &SnapshotStore{db: db, path: dbPath}, nil
}

// Close безопасно закрывает соединение с файлом базы
func (s *SnapshotStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func partKey(id int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b
}

// UpsertOffer сериализует ровно одну деталь в XML и точечно сохраняет в слепок
func (s *SnapshotStore) UpsertOffer(part Part, sellerINN string, baseImageURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := partKey(part.ID)

	// Если деталь помечена на удаление, удаляем её из слепка
	if part.ToDeleteAt != nil {
		return s.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(bucketOffers)
			return b.Delete(key)
		})
	}

	offer := BuildOffer(part, sellerINN, baseImageURL)
	chunk, err := RenderOfferXML(offer)
	if err != nil {
		return fmt.Errorf("сериализация оффера %d: %w", part.ID, err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketOffers)
		return b.Put(key, chunk)
	})
}

// DeleteOffer удаляет деталь из слепка
func (s *SnapshotStore) DeleteOffer(partID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := partKey(partID)
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketOffers)
		return b.Delete(key)
	})
}

// BulkLoad производит первичную быструю загрузку всего каталога в слепок в единой транзакции
func (s *SnapshotStore) BulkLoad(parts []Part, sellerINNs map[int64]string, baseImageURL string, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(bucketOffers); err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return err
		}
		b, err := tx.CreateBucket(bucketOffers)
		if err != nil {
			return err
		}

		count := 0
		for _, part := range parts {
			if part.ToDeleteAt != nil {
				continue
			}
			offer := BuildOffer(part, sellerINNs[part.SellerID], baseImageURL)
			chunk, err := RenderOfferXML(offer)
			if err != nil {
				continue
			}
			if err := b.Put(partKey(part.ID), chunk); err != nil {
				return err
			}
			count++
		}

		meta := tx.Bucket(bucketMeta)
		if err := meta.Put([]byte("inventory_version"), []byte(version)); err != nil {
			return err
		}
		countBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(countBytes, uint64(count))
		return meta.Put([]byte("parts_count"), countBytes)
	})
}

// WriteFullXML собирает и атомарно записывает полный XML-каталог на диск
// из уже готовых байтовых фрагментов за десятки миллисекунд без рефлексии и парсинга
func (s *SnapshotStore) WriteFullXML(targetPath string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, fmt.Errorf("создать директорию для XML: %w", err)
	}

	tmpPath := targetPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("создать временный XML файл: %w", err)
	}

	bw := bufio.NewWriterSize(f, 64*1024)

	// Заголовок документа
	if _, err := bw.WriteString(xml.Header); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return 0, err
	}
	if _, err := bw.WriteString("<offers>\n"); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return 0, err
	}

	count := 0
	err = s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketOffers)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if _, err := bw.Write(v); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return 0, fmt.Errorf("запись фрагментов офферов: %w", err)
	}

	if _, err := bw.WriteString("</offers>\n"); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return 0, err
	}

	if err := bw.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return 0, err
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return 0, err
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return 0, fmt.Errorf("атомарное переименование XML: %w", err)
	}

	return count, nil
}

// GenerateFullXMLBytes собирает полный XML в память (для API и отправки пакетов)
func (s *SnapshotStore) GenerateFullXMLBytes() ([]byte, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString("<offers>\n")

	count := 0
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketOffers)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			buf.Write(v)
			count++
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	buf.WriteString("</offers>\n")
	return buf.Bytes(), count, nil
}

// Count возвращает количество офферов в слепке
func (s *SnapshotStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	_ = s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketOffers)
		if b != nil {
			count = b.Stats().KeyN
		}
		return nil
	})
	return count
}

// GetVersion возвращает сохраненную версию отпечатка склада
func (s *SnapshotStore) GetVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var v string
	_ = s.db.View(func(tx *bbolt.Tx) error {
		meta := tx.Bucket(bucketMeta)
		if meta != nil {
			if val := meta.Get([]byte("inventory_version")); val != nil {
				v = string(val)
			}
		}
		return nil
	})
	return v
}

// SetVersion обновляет отпечаток склада в метаданных слепка
func (s *SnapshotStore) SetVersion(v string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		meta := tx.Bucket(bucketMeta)
		return meta.Put([]byte("inventory_version"), []byte(v))
	})
}
