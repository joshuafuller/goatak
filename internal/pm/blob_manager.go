package pm

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/kdudkov/goatak/pkg/util"
)

var (
	ErrNotFound = errors.New("blob is not found")
	ErrNoHash   = errors.New("no hash")
	ErrBadHash  = errors.New("bad hash")
)

type BlobManager struct {
	logger  *slog.Logger
	mx      sync.RWMutex
	basedir string
}

func NewBlobManages(basedir string) *BlobManager {
	_ = os.MkdirAll(basedir, 0777)

	return &BlobManager{
		logger:  slog.With("logger", "file_manager"),
		mx:      sync.RWMutex{},
		basedir: basedir,
	}
}

func (m *BlobManager) Key(scope, hash string) string {
	if scope == "" {
		return filepath.Join(m.basedir, hash)
	}

	_ = os.MkdirAll(filepath.Join(m.basedir, scope), 0777)

	return filepath.Join(m.basedir, scope, hash)
}

func (m *BlobManager) GetFile(hash string, scope string) (io.ReadSeekCloser, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if hash == "" || !util.FileExists(m.Key(scope, hash)) {
		return nil, ErrNotFound
	}

	return os.Open(m.Key(scope, hash))
}

func (m *BlobManager) Delete(hash string, scope string) error {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if hash == "" || !util.FileExists(m.Key(scope, hash)) {
		return ErrNotFound
	}

	return os.Remove(m.Key(scope, hash))
}

func (m *BlobManager) GetZipFile(hash string, scope string, name string) (*zip.File, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if hash == "" || !util.FileExists(m.Key(scope, hash)) {
		return nil, ErrNotFound
	}

	z, err := zip.OpenReader(m.Key(scope, hash))
	
	if err != nil {
		return nil, err
	}

	for _, x := range z.File {
		if name == x.Name {
			return x, nil
		}
	}

	return nil, errors.New("not found")
}

func (m *BlobManager) ListFiles(hash string, scope string) ([]string, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if hash == "" || !util.FileExists(m.Key(scope, hash)) {
		return nil, ErrNotFound
	}
	
	z, err := zip.OpenReader(m.Key(scope, hash))
	
	if err != nil {
		return nil, err
	}
	
	res := make([]string, len(z.File))
	
	for i, x := range z.File {
		res[i] = x.Name
	}
	
	return res, nil
}

func (m *BlobManager) GetFileStat(scope, hash string) (os.FileInfo, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if hash == "" {
		return nil, ErrNoHash
	}

	return os.Stat(m.Key(scope, hash))
}

func (m *BlobManager) PutFile(scope, hash string, r io.Reader) (string, int64, error) {
	if r == nil {
		return "", 0, errors.New("no reader")
	}

	m.mx.Lock()
	defer m.mx.Unlock()

	if hash != "" && util.FileExists(m.Key(scope, hash)) {
		return hash, 0, nil
	}

	f, err := os.CreateTemp("", "atak-upl-")

	if err != nil {
		return "", 0, err
	}

	h := sha256.New()

	rr := io.TeeReader(r, h)

	var n int64

	if n, err = io.Copy(f, rr); err != nil {
		return "", 0, err
	}

	hash1 := hex.EncodeToString(h.Sum(nil))

	if hash != "" && hash != hash1 {
		return "", 0, ErrBadHash
	}
	
	if err1 := f.Close(); err1 != nil {
		return "", 0, err1
	}

	err = os.Rename(f.Name(), m.Key(scope, hash1))

	return hash1, n, err
}
