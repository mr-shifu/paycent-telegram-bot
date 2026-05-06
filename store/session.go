package store

import "sync"

// Session holds the authenticated state for a Telegram user.
type Session struct {
	PaycentJWT string
	UserID     string
	// pending tracks multi-step conversation state (e.g. waiting for phone, OTP)
	pending pendingState
}

type pendingState int

const (
	PendingNone pendingState = iota
	PendingPhone
	PendingOTP
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[int64]*Session
	// otp scratch pad: maps chatID → phone (awaiting OTP confirmation)
	phoneByChat map[int64]string
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions:    make(map[int64]*Session),
		phoneByChat: make(map[int64]string),
	}
}

func (s *SessionStore) Get(telegramUserID int64) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[telegramUserID]
	return sess, ok
}

func (s *SessionStore) Set(telegramUserID int64, jwt, paycentUserID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[telegramUserID] = &Session{PaycentJWT: jwt, UserID: paycentUserID}
}

func (s *SessionStore) SetPending(telegramUserID int64, state pendingState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[telegramUserID]; ok {
		sess.pending = state
		return
	}
	s.sessions[telegramUserID] = &Session{pending: state}
}

func (s *SessionStore) GetPending(telegramUserID int64) pendingState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sess, ok := s.sessions[telegramUserID]; ok {
		return sess.pending
	}
	return PendingNone
}

func (s *SessionStore) SetPhone(telegramUserID int64, phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phoneByChat[telegramUserID] = phone
}

func (s *SessionStore) GetPhone(telegramUserID int64) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	phone, ok := s.phoneByChat[telegramUserID]
	return phone, ok
}

func (s *SessionStore) ClearPhone(telegramUserID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.phoneByChat, telegramUserID)
}
