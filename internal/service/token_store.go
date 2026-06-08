package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]int64
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]int64),
	}
}

func (s *TokenStore) Create(userID int64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(bytes)

	s.mu.Lock()
	s.tokens[token] = userID
	s.mu.Unlock()

	return token, nil
}

func (s *TokenStore) GetUserID(token string) (int64, bool) {
	s.mu.RLock()
	userID, ok := s.tokens[token]
	s.mu.RUnlock()

	return userID, ok
}

func (s *TokenStore) Delete(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}
