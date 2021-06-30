package poll

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	// "errors"
	"github.com/lllamnyp/consensus-backend/internal/speakers"
	"time"
)

type Poll interface {
	AddAnswer(User, Answer)
	ListAnswers() []Answer
	Respond(User, Answer)
	ToggleVote(User, Answer)
	GetAnswerByID(string) (Answer, error)
}

type Answer interface {
	GetID() string
	GetAsker() User
	GetAddressee() User
	GetContent() string
	GetResponse() string
	GetWithUser() User
	GetTimestamp() int
	HasVoted(u User) bool
	WithUser(User)
	WithResponse(string)
	WithVoters(map[string]User)
	ToggleVote(u User)
	MarshalJSON() ([]byte, error)
}

type User interface {
	GetName() string
	GetLogin() string
	GetID() string
}

type Store interface {
	AddAnswer(User, Answer)
	ListAnswers() []Answer
	Respond(User, Answer)
	ToggleVote(User, Answer)
	GetAnswerByID(string) (Answer, error)
}

type poll struct {
	store Store
}

func (p *poll) AddAnswer(u User, a Answer) {
	p.store.AddAnswer(u, a)
}

func (p *poll) Respond(u User, a Answer) {
	p.store.Respond(u, a)
}

func (p *poll) ListAnswers() []Answer {
	return p.store.ListAnswers()
}

func (p *poll) ToggleVote(u User, a Answer) {
	p.store.ToggleVote(u, a)
}

func (p *poll) GetAnswerByID(id string) (Answer, error) {
	return p.store.GetAnswerByID(id)
}

type answer struct {
	id        string
	asker     User
	addressee User
	content   string
	response  string
	voters    map[string]User
	timestamp int
	withUser  User
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

func (a *answer) GetResponse() string {
	return a.response
}

func (a *answer) GetAsker() User {
	return a.asker
}

func (a *answer) GetAddressee() User {
	return a.addressee
}

func (a *answer) GetWithUser() User {
	return a.withUser
}

func (a *answer) GetTimestamp() int {
	return a.timestamp
}

func (a *answer) WithUser(u User) {
	a.withUser = u
}

func (a *answer) WithResponse(s string) {
	a.response = s
}

func (a *answer) GetID() string {
	return a.id
}

func (a *answer) ToggleVote(u User) {
	if _, ok := a.voters[u.GetID()]; !ok {
		a.voters[u.GetID()] = u
		return
	}
	delete(a.voters, u.GetID())
}

func (a *answer) WithVoters(v map[string]User) {
	a.voters = v
}

func (a *answer) MarshalJSON() ([]byte, error) {
	var upvoted bool = a.HasVoted(a.withUser)
	type marshalableAnswer struct {
		ID        string `json:"id"`
		Asker     string `json:"asker"`
		Addressee int    `json:"addressee"`
		Content   string `json:"content"`
		Response  string `json:"response"`
		Upvoted   bool   `json:"upvoted"`
		Votes     int    `json:"votes"`
		Timestamp int    `json:"timestamp"`
	}
	return json.Marshal(marshalableAnswer{a.id, a.asker.GetName(), speakers.Lookup(a.addressee.GetLogin()), a.content, a.response, upvoted, len(a.voters), a.timestamp})
}

func New(s Store) Poll {
	return &poll{s}
}

/*
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
*/
func NewAnswer(content string, asker User, addressee User, anon bool) Answer {
	ask := AnonymousUser()
	add := AnonymousUser()
	if addressee != nil {
		add = addressee
	}
	votes := map[string]User{}
	if !anon {
		ask = asker
		votes[asker.GetID()] = asker
	}
	return &answer{hash(content), ask, add, content, "", votes, int(time.Now().Unix()), nil}
}

type user struct {
	login string
	name  string
	id    string
}

func AnonymousUser() User {
	return &user{"", "", ""}
}

func NewUser(name, login string) User {
	return &user{login, name, hash(login)}
}

func (u *user) GetName() string {
	return u.name
}

func (u *user) GetID() string {
	return u.id
}

func (u *user) GetLogin() string {
	return u.login
}

func hash(s string) string {
	hasher := sha1.New()
	hasher.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
