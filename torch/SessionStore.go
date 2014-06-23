package torch

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type SessionStore struct {
	sessions  map[string]*Session
	Broadcast func(name string, val interface{})
}

func InMemorySessionStore() *SessionStore {
	ss := SessionStore{
		sessions: make(map[string]*Session),
	}

	go ss.StartExpiry()
	return &ss
}

func (ss *SessionStore) DumpSessions(w io.Writer) {
	for _, session := range ss.sessions {
		if session.Key != nil && session.User != nil {
			w.Write([]byte(fmt.Sprintf("%s|%d\n", *session.Key, session.User.Id)))
		}
	}
}

func (ss *SessionStore) LoadSessions(r io.Reader, loadUser func(uint64) (*User, error)) {
	lr := bufio.NewReader(r)
	for {
		line, err := lr.ReadString('\n')
		if err != nil {
			break
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		userId, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			log.Println("Error loading session: %s\n", err)
			continue
		}
		s := &Session{
			Key:         &parts[0],
			Store:       ss,
			LastRequest: time.Now(),
		}
		user, err := loadUser(userId)
		if err != nil {
			log.Println("Error loading session: %s\n", err)
			continue
		}
		s.User = user
		log.Printf("Hidrated session for user %d\n", user.Id)
		ss.sessions[*s.Key] = s

	}
}

func (ss *SessionStore) StartExpiry() {
	for {
		time.Sleep(time.Second * 10)

		log.Println("CHECK SESSION EXPIRY")
		for key, s := range ss.sessions {
			if time.Since(s.LastRequest).Minutes() > 30 {
				log.Printf("Expire Session %s", key)
				delete(ss.sessions, key)
			}
		}
	}

}

func (ss *SessionStore) NewSession() *Session {
	randBytes := make([]byte, 128, 128)
	_, _ = rand.Reader.Read(randBytes)
	keyString := hex.EncodeToString(randBytes)
	session := Session{
		Key:         &keyString,
		Store:       ss,
		LastRequest: time.Now(),
	}
	ss.sessions[keyString] = &session
	return &session
}

func (ss *SessionStore) GetSession(key string) (*Session, error) {
	sess, ok := ss.sessions[key]
	if !ok {
		fmt.Printf("Session Not Found: %s\n", key)
		return nil, errors.New("No session with that key")
	}
	sess.LastRequest = time.Now()
	return sess, nil
}