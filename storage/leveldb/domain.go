package database

import (
	"fmt"

	"git.happydns.org/happydns/model"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (s *LevelDBStorage) GetDomains(u *happydns.User) (domains happydns.Domains, err error) {
	iter := s.search("domain-")
	defer iter.Release()

	for iter.Next() {
		var z happydns.Domain

		err = decodeData(iter.Value(), &z)
		if err != nil {
			return
		}

		if z.IdUser == u.Id {
			domains = append(domains, &z)
		}
	}

	return
}

func (s *LevelDBStorage) GetDomain(u *happydns.User, id int) (z *happydns.Domain, err error) {
	z = &happydns.Domain{}
	err = s.get(fmt.Sprintf("domain-%d", id), &z)

	if z.IdUser != u.Id {
		z = nil
		err = leveldb.ErrNotFound
	}

	return
}

func (s *LevelDBStorage) GetDomainByDN(u *happydns.User, dn string) (*happydns.Domain, error) {
	domains, err := s.GetDomains(u)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if domain.DomainName == dn {
			return domain, nil
		}
	}

	return nil, leveldb.ErrNotFound
}

func (s *LevelDBStorage) DomainExists(dn string) bool {
	iter := s.search("domain-")
	defer iter.Release()

	for iter.Next() {
		var z happydns.Domain

		err := decodeData(iter.Value(), &z)
		if err != nil {
			continue
		}

		if z.DomainName == dn {
			return true
		}
	}

	return false
}

func (s *LevelDBStorage) CreateDomain(u *happydns.User, z *happydns.Domain) error {
	key, id, err := s.findInt63Key("domain-")
	if err != nil {
		return err
	}

	z.Id = id
	z.IdUser = u.Id
	return s.put(key, z)
}

func (s *LevelDBStorage) UpdateDomain(z *happydns.Domain) error {
	return s.put(fmt.Sprintf("domain-%d", z.Id), z)
}

func (s *LevelDBStorage) UpdateDomainOwner(z *happydns.Domain, newOwner *happydns.User) error {
	z.IdUser = newOwner.Id
	return s.put(fmt.Sprintf("domain-%d", z.Id), z)
}

func (s *LevelDBStorage) DeleteDomain(z *happydns.Domain) error {
	return s.delete(fmt.Sprintf("domain-%d", z.Id))
}

func (s *LevelDBStorage) ClearDomains() error {
	tx, err := s.db.OpenTransaction()
	if err != nil {
		return err
	}

	iter := tx.NewIterator(util.BytesPrefix([]byte("domain-")), nil)
	defer iter.Release()

	for iter.Next() {
		err = tx.Delete(iter.Key(), nil)
		if err != nil {
			tx.Discard()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Discard()
		return err
	}

	return nil
}