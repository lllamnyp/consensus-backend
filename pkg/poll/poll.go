package poll

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
)

type Poll interface {
	AddAnswer(User, Answer)
	ListAnswers() map[string]Answer
	ToggleVote(User, Answer)
	GetAnswerByID(string) (Answer, error)
}

type Answer interface {
	HasVoted(u User) bool
	GetContent() string
	GetID() string
	WithUser(User)
	WithVoters(map[string]User)
	GetWithUser() User
	ToggleVote(u User)
	MarshalJSON() ([]byte, error)
}

type User interface {
	GetName() string
	GetID() string
}

type Store interface {
	AddAnswer(User, Answer)
	ListAnswers() map[string]Answer
	ToggleVote(User, Answer)
	GetAnswerByID(string) (Answer, error)
}

type poll struct {
	store Store
}

func (p *poll) AddAnswer(u User, a Answer) {
	p.store.AddAnswer(u, a)
}

func (p *poll) ListAnswers() map[string]Answer {
	return p.store.ListAnswers()
}

func (p *poll) ToggleVote(u User, a Answer) {
	p.store.ToggleVote(u, a)
}

func (p *poll) GetAnswerByID(id string) (Answer, error) {
	return p.store.GetAnswerByID(id)
}

type answer struct {
	content  string
	id       string
	voters   map[string]User
	withUser User
}

func (a *answer) HasVoted(u User) bool {
	if u == nil {
		return false
	}
	_, ok := a.voters[u.GetID()]
	return ok
}

func (a *answer) GetContent() string {
	return a.content
}

func (a *answer) GetID() string {
	return a.id
}

func (a *answer) WithUser(u User) {
	a.withUser = u
}

func (a *answer) WithVoters(v map[string]User) {
	a.voters = v
}

func (a *answer) GetWithUser() User {
	return a.withUser
}

func (a *answer) ToggleVote(u User) {
	if _, ok := a.voters[u.GetID()]; ok {
		delete(a.voters, u.GetID())
	} else {
		a.voters[u.GetID()] = u
	}
}

func (a *answer) MarshalJSON() ([]byte, error) {
	var upvoted bool = a.HasVoted(a.withUser)
	type marshalableAnswer struct {
		Content string `json:"content"`
		Upvoted bool   `json:"upvoted"`
		Votes   int    `json:"votes"`
	}
	return json.Marshal(marshalableAnswer{a.content, upvoted, len(a.voters)})
}

func New(s Store) Poll {
	return &poll{s}
}

func NewInMemoryStore() Store {
	return &store{make(map[string]Answer)}
}

type store struct {
	answers map[string]Answer
}

func (s *store) AddAnswer(u User, a Answer) {
	_, ok := s.answers[a.GetID()]
	if ok {
		if s.answers[a.GetID()].HasVoted(u) {
			return
		}
	} else {
		s.answers[a.GetID()] = a
	}
	s.ToggleVote(u, a)
}

func (s *store) ListAnswers() map[string]Answer {
	return s.answers
}

func (s *store) ToggleVote(u User, a Answer) {
	a.ToggleVote(u)
}

func (s *store) GetAnswerByID(id string) (Answer, error) {
	if a, ok := s.answers[id]; ok {
		return a, nil
	}
	return nil, errors.New("No answer with id:" + id)
}

func NewAnswer(content string) Answer {
	return &answer{content, hash(content), map[string]User{}, nil}
}

type user struct {
	name string
	id   string
}

func NewUser(name string) User {
	return &user{name, hash(name)}
}

func (u *user) GetName() string {
	return u.name
}

func (u *user) GetID() string {
	return u.id
}

func hash(s string) string {
	hasher := sha1.New()
	hasher.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
