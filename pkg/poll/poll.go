package poll

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
)

type Poll interface {
	AddAnswer(User, Answer)
	ListAnswers() map[string]Answer
	GetAnswerByID(string) Answer
	ToggleVote(User, Answer)
}

type Answer interface {
	HasVoted(u User) bool
	GetContent() string
	GetID() string
	MarshalJSON() ([]byte, error)
}

type User interface {
	GetName() string
	GetID() string
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

func (p *poll) GetAnswerByID(string) Answer {
	return &answer{}
}

func (p *poll) ToggleVote(User, Answer) {}

type answer struct {
	content string
	id      string
	voters  map[string]User
}

func (a *answer) HasVoted(u User) bool {
	_, ok := a.voters[u.GetID()]
	return ok
}

func (a *answer) GetContent() string {
	return a.content
}

func (a *answer) GetID() string {
	return a.id
}

func (a *answer) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.content)
}

type Store interface {
	AddAnswer(User, Answer)
	ListAnswers() map[string]Answer
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
	s.answers[a.GetID()] = a
}

func (s *store) ListAnswers() map[string]Answer {
	return s.answers
}

func NewAnswer(content string) Answer {
	return &answer{content, hash(content), map[string]User{}}
}

func hash(s string) string {
	hasher := sha1.New()
	hasher.Write([]byte(s))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
