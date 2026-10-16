package repository

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

	"export-service/internal/domain"
	"go.etcd.io/bbolt"
)

var (
	bucketOffers = []byte("offers")
	bucketMeta   = []byte("meta")
)

// SnapshotRepository определяет контракт хранилища инкрементального слепка каталога
type SnapshotRepository interface {
	SaveOffer(part domain.Part, sellerINN string, baseImageURL string) error
	DeleteOffer(partID int64) error
	BulkLoad(parts []domain.Part, sellerINNs map[int64]string, baseImageURL string, version string) error
	WriteFullXML(targetPath string) (int, error)
	GenerateFullXMLBytes() ([]byte, int, error)
	Count() int
	GetVersion() string
	SetVersion(version string) error
	Close() error
}

// BoltSnapshotRepository реализует хранение слепка в локальной базе bbolt
type BoltSnapshotRepository struct {
	db   *bbolt.DB
	path string
	mu   sync.RWMutex
}

// NewBoltSnapshotRepository открывает или создает файл базы bbolt
func NewBoltSnapshotRepository(dbPath string) (*BoltSnapshotRepository, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("создать директорию базы слепка: %w", err)
	}

	db, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("открыть bbolt базу (%s): %w", dbPath, err)
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

	return &BoltSnapshotRepository{db: db, path: dbPath}, nil
}

// Close закрывает базу данных
func (r *BoltSnapshotRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func encodePartKey(id int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b
}

// SaveOffer точечно обновляет или удаляет оффер детали в слепке
func (r *BoltSnapshotRepository) SaveOffer(part domain.Part, sellerINN string, baseImageURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := encodePartKey(part.ID)

	if part.ToDeleteAt != nil {
		return r.db.Update(func(tx *bbolt.Tx) error {
			return tx.Bucket(bucketOffers).Delete(key)
		})
	}

	offer := domain.BuildOffer(part, sellerINN, baseImageURL)
	chunk, err := domain.RenderOfferXML(offer)
	if err != nil {
		return fmt.Errorf("сериализация оффера детали %d: %w", part.ID, err)
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketOffers).Put(key, chunk)
	})
}

// DeleteOffer удаляет оффер детали из базы слепка
func (r *BoltSnapshotRepository) DeleteOffer(partID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := encodePartKey(partID)
	return r.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketOffers).Delete(key)
	})
}

// BulkLoad производит первичную быструю загрузку всех запчастей в слепок в единой транзакции
func (r *BoltSnapshotRepository) BulkLoad(parts []domain.Part, sellerINNs map[int64]string, baseImageURL string, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.db.Update(func(tx *bbolt.Tx) error {
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
			offer := domain.BuildOffer(part, sellerINNs[part.SellerID], baseImageURL)
			chunk, err := domain.RenderOfferXML(offer)
			if err != nil {
				continue
			}
			if err := b.Put(encodePartKey(part.ID), chunk); err != nil {
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

// WriteFullXML выполняет потоковую запись полного XML на диск через линейный курсор bbolt
func (r *BoltSnapshotRepository) WriteFullXML(targetPath string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, fmt.Errorf("создать директорию: %w", err)
	}

	tmpPath := targetPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("создать временный XML файл: %w", err)
	}

	bw := bufio.NewWriterSize(f, 64*1024)

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
	err = r.db.View(func(tx *bbolt.Tx) error {
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
		return 0, fmt.Errorf("запись офферов: %w", err)
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
		return 0, fmt.Errorf("атомарное переименование: %w", err)
	}

	return count, nil
}

// GenerateFullXMLBytes генерирует полный XML-документ в памяти
func (r *BoltSnapshotRepository) GenerateFullXMLBytes() ([]byte, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString("<offers>\n")

	count := 0
	err := r.db.View(func(tx *bbolt.Tx) error {
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

// Count возвращает количество офферов в базе
func (r *BoltSnapshotRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	_ = r.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketOffers)
		if b != nil {
			count = b.Stats().KeyN
		}
		return nil
	})
	return count
}

// GetVersion возвращает текущий отпечаток склада из метаданных
func (r *BoltSnapshotRepository) GetVersion() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var v string
	_ = r.db.View(func(tx *bbolt.Tx) error {
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

// SetVersion обновляет отпечаток склада
func (r *BoltSnapshotRepository) SetVersion(version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.db.Update(func(tx *bbolt.Tx) error {
		meta := tx.Bucket(bucketMeta)
		return meta.Put([]byte("inventory_version"), []byte(version))
	})
}
