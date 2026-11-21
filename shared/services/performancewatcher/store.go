package performancewatcher

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Store struct {
	db   *leveldb.DB
	path string
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}
	db, err := leveldb.OpenFile(path, &opt.Options{Strict: opt.StrictAll})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	v, err := db.Get([]byte("schema"), nil)
	if err == leveldb.ErrNotFound {
		err = db.Put([]byte("schema"), []byte("1"), &opt.WriteOptions{Sync: true})
	} else if err == nil && string(v) != "1" {
		err = fmt.Errorf("unsupported watcher database schema %q", v)
	}
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func epochKey(epoch uint64) []byte {
	k := make([]byte, 9)
	k[0] = 'e'
	binary.BigEndian.PutUint64(k[1:], epoch)
	return k
}

func (s *Store) Registry() ([]Validator, error) {
	data, err := s.db.Get([]byte("registry"), nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var v []Validator
	err = json.Unmarshal(data, &v)
	return v, err
}

// Ordinals are append-only, so later discovery cannot inherit earlier coverage.
func (s *Store) UpdateRegistry(in []Validator) ([]Validator, error) {
	v, err := s.Registry()
	if err != nil {
		return nil, err
	}
	positions := make(map[uint64]int, len(v))
	for i := range v {
		positions[v[i].Index] = i
		v[i].Eligible = false
	}
	for _, item := range in {
		if i, ok := positions[item.Index]; ok {
			if v[i].Pubkey != item.Pubkey {
				return nil, fmt.Errorf("validator identity changed at index %d", item.Index)
			}
			v[i] = item
		} else {
			positions[item.Index] = len(v)
			v = append(v, item)
		}
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	err = s.db.Put([]byte("registry"), data, &opt.WriteOptions{Sync: true})
	return v, err
}

func encodeRecord(r EpochRecord) ([]byte, error) {
	if r.Count < 0 || uint64(r.Count) > uint64(^uint32(0)) || len(r.Bits) != (r.Count+3)/4 {
		return nil, fmt.Errorf("invalid epoch record")
	}
	v := make([]byte, 68+len(r.Bits))
	binary.BigEndian.PutUint32(v, uint32(r.Count))
	for i, root := range []string{r.TargetRoot, r.FinalizedRoot} {
		b, err := hex.DecodeString(strings.TrimPrefix(root, "0x"))
		if err != nil || len(b) != 32 {
			return nil, fmt.Errorf("invalid provenance root")
		}
		copy(v[4+i*32:], b)
	}
	copy(v[68:], r.Bits)
	return v, nil
}
func decodeRecord(epoch uint64, v []byte) (EpochRecord, error) {
	if len(v) < 68 {
		return EpochRecord{}, fmt.Errorf("corrupt watcher epoch %d", epoch)
	}
	n := int(binary.BigEndian.Uint32(v))
	if len(v) != 68+(n+3)/4 {
		return EpochRecord{}, fmt.Errorf("corrupt watcher epoch %d length", epoch)
	}
	return EpochRecord{Epoch: epoch, Count: n, TargetRoot: "0x" + hex.EncodeToString(v[4:36]), FinalizedRoot: "0x" + hex.EncodeToString(v[36:68]), Bits: bytes.Clone(v[68:])}, nil
}
func (s *Store) Put(r EpochRecord) error {
	if data, err := s.db.Get(epochKey(r.Epoch), nil); err == nil {
		previous, err := decodeRecord(r.Epoch, data)
		if err != nil {
			return err
		}
		if previous.TargetRoot != r.TargetRoot {
			return ErrFinalityConflict
		}
		for i := 0; i < min(previous.Count, r.Count); i++ {
			a, b := previous.Get(i), r.Get(i)
			if (a == Timely || a == Missed) && (b == Timely || b == Missed) && a != b {
				return fmt.Errorf("epoch %d validator ordinal %d changed finalized observation: %w", r.Epoch, i, ErrFinalityConflict)
			}
		}
	} else if err != leveldb.ErrNotFound {
		return err
	}
	v, err := encodeRecord(r)
	if err != nil {
		return err
	}
	// Coverage is the count and packed observations in this same atomic record.
	batch := new(leveldb.Batch)
	batch.Put(epochKey(r.Epoch), v)
	return s.db.Write(batch, &opt.WriteOptions{Sync: true})
}
func (s *Store) Scan(oldest uint64, f func(EpochRecord) error) error {
	snap, err := s.db.GetSnapshot()
	if err != nil {
		return err
	}
	defer snap.Release()
	it := snap.NewIterator(&util.Range{Start: epochKey(oldest), Limit: []byte{'f'}}, nil)
	defer it.Release()
	for it.Next() {
		if len(it.Key()) != 9 {
			return fmt.Errorf("corrupt watcher epoch key")
		}
		r, err := decodeRecord(binary.BigEndian.Uint64(it.Key()[1:]), it.Value())
		if err != nil {
			return err
		}
		if err = f(r); err != nil {
			return err
		}
	}
	return it.Error()
}
func (s *Store) Prune(oldest uint64, validators []Validator) error {
	batch := new(leveldb.Batch)
	err := s.Scan(0, func(r EpochRecord) error {
		if r.Epoch < oldest {
			batch.Delete(epochKey(r.Epoch))
		} else {
			changed := false
			for i := 0; i < r.Count; i++ {
				if i >= len(validators) {
					return fmt.Errorf("epoch references missing registry ordinal")
				}
				v := validators[i]
				if (!v.Eligible || r.Epoch < v.Activation) && r.Get(i) != Unknown {
					r.Set(i, Unknown)
					changed = true
				}
			}
			if changed {
				data, err := encodeRecord(r)
				if err != nil {
					return err
				}
				batch.Put(epochKey(r.Epoch), data)
			}
		}
		if batch.Len() >= 256 {
			if err := s.db.Write(batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.db.Write(batch, &opt.WriteOptions{Sync: true})
}
func (s *Store) Compact(oldest uint64) error {
	return s.db.CompactRange(util.Range{Start: []byte{'e'}, Limit: epochKey(oldest)})
}
func (s *Store) PruneBefore(oldest uint64) error {
	it := s.db.NewIterator(&util.Range{Start: []byte{'e'}, Limit: epochKey(oldest)}, nil)
	defer it.Release()
	batch := new(leveldb.Batch)
	for it.Next() {
		batch.Delete(bytes.Clone(it.Key()))
		if batch.Len() >= 256 {
			if err := s.db.Write(batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}
	}
	if err := it.Error(); err != nil {
		return err
	}
	return s.db.Write(batch, &opt.WriteOptions{Sync: true})
}
func (s *Store) Size() uint64 {
	var n uint64
	_ = filepath.WalkDir(s.path, func(_ string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if info, e := d.Info(); e == nil {
				n += uint64(info.Size())
			}
		}
		return nil
	})
	return n
}
