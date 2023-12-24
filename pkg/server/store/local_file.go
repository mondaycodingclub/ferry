package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type LocalFileStore struct {
	sync.Mutex
}

func (s *LocalFileStore) Find(key string) (interface{}, error) {
	s.Lock()
	defer s.Unlock()

	data, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("not found")
	}
	var value any
	err = json.Unmarshal(data, &value)
	if err != nil {
		return nil, fmt.Errorf("")
	}
	return value, nil
}

func (s *LocalFileStore) Save(key string, value interface{}) error {
	s.Lock()
	defer s.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(key, data, 0644)
}

func (s *LocalFileStore) Delete(key string) error {
	return os.RemoveAll(key)
}
