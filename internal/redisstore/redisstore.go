package redisstore

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
)

func NewRedisStore(c *redis.Client) poll.Store {
	c.LPush("answerids")
	return &redisstore{c}
}

type redisstore struct {
	client *redis.Client
}

func (s *redisstore) AddAnswer(u poll.User, a poll.Answer) {
	s.client.LPush("answerids", a.GetID())
	s.client.HMSet("answer:"+a.GetID(), map[string]interface{}{
		"id":        a.GetID(),
		"asker":     a.GetAsker().GetID(),
		"addressee": a.GetAddressee().GetID(),
		"content":   a.GetContent(),
		"response":  a.GetResponse(),
		"timestamp": strconv.Itoa(a.GetTimestamp()),
	})
	if u.GetID() != "" {
		s.client.SAdd("voters:"+a.GetID(), u.GetID())
		s.client.HMSet("user:"+u.GetID(), map[string]interface{}{"login": u.GetLogin(), "name": u.GetName()})
		s.client.HMSet("user:"+a.GetAddressee().GetID(), map[string]interface{}{"login": a.GetAddressee().GetLogin(), "name": a.GetAddressee().GetName()})
	}
	return
}

func (s *redisstore) ListAnswers() []poll.Answer {
	answerids := s.client.LRange("answerids", 0, -1).Val()
	answers := make([]poll.Answer, 0, 8)
	for _, id := range answerids {
		voterids := s.client.SMembers("voters:" + id).Val()
		voters := make(map[string]poll.User)
		for _, voterid := range voterids {
			voter := s.client.HGetAll("user:" + voterid).Val()
			voters[voterid] = poll.NewUser(voter["login"], voter["name"])
		}
		redisAnswer := s.client.HGetAll("answer:" + id).Val()

		var asker poll.User
		if redisAnswer["asker"] == "" {
			asker = poll.AnonymousUser()
		} else {
			redisAsker := s.client.HGetAll("user:" + redisAnswer["asker"]).Val()
			asker = poll.NewUser(redisAsker["login"], redisAsker["name"])
		}

		var addressee poll.User
		if redisAnswer["addressee"] == "" {
			addressee = poll.AnonymousUser()
		} else {
			redisAddressee := s.client.HGetAll("user:" + redisAnswer["addressee"]).Val()
			addressee = poll.NewUser(redisAddressee["login"], redisAddressee["name"])
		}
		anon := asker.GetID() == ""

		answer := poll.NewAnswer(redisAnswer["content"], asker, addressee, anon)
		answer.WithVoters(voters)
		answer.WithResponse(redisAnswer["response"])
		answers = append(answers, answer)
	}
	return answers
}

func (s *redisstore) GetAnswerByID(id string) (poll.Answer, error) {
	exists := s.client.Exists("answer:" + id).Val()
	if exists != 1 {
		return nil, errors.New("No answer with id:" + id)
	}
	voterids := s.client.SMembers("voters:" + id).Val()
	voters := make(map[string]poll.User)
	for _, voterid := range voterids {
		voter := s.client.HGetAll("user:" + voterid).Val()
		voters[voterid] = poll.NewUser(voter["login"], voter["name"])
	}
	redisAnswer := s.client.HGetAll("answer:" + id).Val()

	var asker poll.User
	if redisAnswer["asker"] == "" {
		asker = poll.AnonymousUser()
	} else {
		redisAsker := s.client.HGetAll("user:" + redisAnswer["asker"]).Val()
		asker = poll.NewUser(redisAsker["login"], redisAsker["name"])
	}

	var addressee poll.User
	if redisAnswer["addressee"] == "" {
		addressee = poll.AnonymousUser()
	} else {
		redisAddressee := s.client.HGetAll("user:" + redisAnswer["addressee"]).Val()
		addressee = poll.NewUser(redisAddressee["login"], redisAddressee["name"])
	}
	anon := asker.GetID() == ""

	answer := poll.NewAnswer(redisAnswer["content"], asker, addressee, anon)
	answer.WithVoters(voters)
	answer.WithResponse(redisAnswer["response"])

	return answer, nil
}

func (s *redisstore) Respond(u poll.User, a poll.Answer) {
	/*exists := s.client.Exists("answers:" + id).Val()
	if exists != 1 {
		return nil, errors.New("No answer with id:" + id)
	}*/
	s.client.HMSet("answer:"+a.GetID(), map[string]interface{}{"response": a.GetResponse()})
}

func (s *redisstore) ToggleVote(u poll.User, a poll.Answer) {
	fmt.Printf("Toggle vote for answer id %s with user id %s.\n", a.GetID(), u.GetID())
	exists := s.client.SIsMember("voters:"+a.GetID(), u.GetID()).Val()
	if exists {
		fmt.Printf("Remove vote for answer id %s with user id %s.\n", a.GetID(), u.GetID())
		s.client.SRem("voters:"+a.GetID(), u.GetID())
		return
	}
	result := s.client.SAdd("voters:"+a.GetID(), u.GetID()).Val()
	fmt.Printf("Add vote for answer id %s with user id %s. SADD == %d.\n", a.GetID(), u.GetID(), result)
	s.client.HMSet("user:"+u.GetID(), map[string]interface{}{"login": u.GetLogin(), "name": u.GetName()})
}
