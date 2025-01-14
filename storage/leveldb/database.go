// Copyright or © or Copr. happyDNS (2020)
//
// contact@happydomain.org
//
// This software is a computer program whose purpose is to provide a modern
// interface to interact with DNS systems.
//
// This software is governed by the CeCILL license under French law and abiding
// by the rules of distribution of free software.  You can use, modify and/or
// redistribute the software under the terms of the CeCILL license as
// circulated by CEA, CNRS and INRIA at the following URL
// "http://www.cecill.info".
//
// As a counterpart to the access to the source code and rights to copy, modify
// and redistribute granted by the license, users are provided only with a
// limited warranty and the software's author, the holder of the economic
// rights, and the successive licensors have only limited liability.
//
// In this respect, the user's attention is drawn to the risks associated with
// loading, using, modifying and/or developing or reproducing the software by
// the user in light of its specific status of free software, that may mean
// that it is complicated to manipulate, and that also therefore means that it
// is reserved for developers and experienced professionals having in-depth
// computer knowledge. Users are therefore encouraged to load and test the
// software's suitability as regards their requirements in conditions enabling
// the security of their systems and/or data to be ensured and, more generally,
// to use and operate it in the same conditions as regards security.
//
// The fact that you are presently reading this means that you have had
// knowledge of the CeCILL license and that you accept its terms.

package database // import "happydns.org/storage/leveldb"

import (
	"encoding/json"
	"fmt"
	"log"

	"git.happydns.org/happydomain/model"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LevelDBStorage struct {
	db *leveldb.DB
}

// NewMySQLStorage establishes the connection to the database
func NewLevelDBStorage(path string) (s *LevelDBStorage, err error) {
	var db *leveldb.DB

	db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		if _, ok := err.(*errors.ErrCorrupted); ok {
			log.Println("LevelDB was corrupted; attempting recovery (%s)", err.Error())
			_, err = leveldb.RecoverFile(path, nil)
			if err != nil {
				return
			}
			log.Println("LevelDB recovery succeeded!")
		} else {
			return
		}
	}

	s = &LevelDBStorage{db}
	return
}

func (s *LevelDBStorage) Tidy() error {
	for _, tidy := range []func() error{s.TidySessions, s.TidyAuthUsers, s.TidyUsers, s.TidyProviders, s.TidyDomains, s.TidyZones} {
		if err := tidy(); err != nil {
			return err
		}
	}
	return nil
}

func (s *LevelDBStorage) Close() error {
	return s.db.Close()
}

func decodeData(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s *LevelDBStorage) get(key string, v interface{}) error {
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return err
	}

	return decodeData(data, v)
}

func (s *LevelDBStorage) put(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return s.db.Put([]byte(key), data, nil)
}

func (s *LevelDBStorage) findIdentifierKey(prefix string) (key string, id happydns.Identifier, err error) {
	found := true
	for found {
		id, err = happydns.NewRandomIdentifier()
		if err != nil {
			return
		}
		key = fmt.Sprintf("%s%s", prefix, id.String())

		found, err = s.db.Has([]byte(key), nil)
		if err != nil {
			return
		}
	}
	return
}

func (s *LevelDBStorage) delete(key string) error {
	return s.db.Delete([]byte(key), nil)
}

func (s *LevelDBStorage) search(prefix string) iterator.Iterator {
	return s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
}
