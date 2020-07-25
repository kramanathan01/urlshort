package persist

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"

	"github.com/dgraph-io/badger/v2"
)

type Short struct {
	Path  string
	Site  string
	Count int
}

type database struct {
	DB   *badger.DB
	file string
	opts badger.Options
}

// Db Struct that holds database handle
var Db database

func (db *database) Open() {
	var err error
	db.file = "badger"
	if db.file == "" {
		db.opts = badger.DefaultOptions("").WithInMemory(true)
	} else {
		db.opts = badger.DefaultOptions(os.TempDir() + db.file)
	}
	db.opts.Logger = nil
	db.DB, err = badger.Open(db.opts)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
}

func (s *Short) gobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil

}

func gobDecode(d []byte) (*Short, error) {
	var s *Short
	buf := bytes.NewBuffer(d)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Saves key value to DB
func (db *database) SaveMap(m map[string]Short) error {
	txn := db.DB.NewTransaction(true)
	for k, v := range m {
		gb, _ := v.gobEncode()
		if err := txn.Set([]byte(k), gb); err == badger.ErrTxnTooBig {
			_ = txn.Commit()
			txn = db.DB.NewTransaction(true)
			_ = txn.Set([]byte(k), gb)
		}
	}
	return txn.Commit()
}

// Save single key to DB
func (db *database) Save(s Short) error {
	err := db.DB.Update(func(txn *badger.Txn) error {
		gb, _ := s.gobEncode()
		return txn.Set([]byte(s.Path), []byte(gb))
	})
	return err
}

// Get single key from DB
func (db *database) Get(k string) (*Short, bool) {
	var tr *Short
	err := db.DB.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		data, _ := i.ValueCopy(nil)
		tr, err = gobDecode(data)
		return err
	})
	if err != nil {
		return tr, false
	}
	return tr, true
}

// Get all data from DB
func (db *database) GetAll() (map[string]Short, bool) {
	var ga = make(map[string]Short)
	err := db.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				vv, _ := gobDecode(v)
				ga[string(k)] = *vv
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ga, false
	}
	return ga, true
}
