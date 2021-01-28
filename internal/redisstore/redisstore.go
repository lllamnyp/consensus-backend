package redisstore

import (
	"errors"

	"github.com/go-redis/redis"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
)

func NewRedisStore(c *redis.Client) poll.Store {
	c.HMSet("answers", map[string]interface{}{})
	return &redisstore{c}
}

type redisstore struct {
	client *redis.Client
}

func (s *redisstore) AddAnswer(u poll.User, a poll.Answer) {
	s.client.HMSet("answers", map[string]interface{}{a.GetID(): a.GetContent()})
	s.client.HMSet("voters:"+a.GetID(), map[string]interface{}{u.GetID(): u.GetName()})
	return
}

func (s *redisstore) ListAnswers() map[string]poll.Answer {
	answers := make(map[string]poll.Answer)
	answerHM := s.client.HGetAll("answers").Val()
	for answerID, answerContent := range answerHM {
		votersHM := s.client.HGetAll("voters:" + answerID).Val()
		voters := make(map[string]poll.User)
		for userID, userName := range votersHM {
			voters[userID] = poll.NewUser(userName)
		}
		answers[answerID] = poll.NewAnswer(answerContent)
		answers[answerID].WithVoters(voters)
	}
	return answers
}

func (s *redisstore) GetAnswerByID(id string) (poll.Answer, error) {
	exists := s.client.HExists("answers", id).Val()
	if !exists {
		return nil, errors.New("No answer with id:" + id)
	}
	answer := poll.NewAnswer(s.client.HGet("answers", id).Val())
	return answer, nil
}

func (s *redisstore) ToggleVote(u poll.User, a poll.Answer) {
	exists := s.client.HExists("voters:"+a.GetID(), u.GetID()).Val()
	if exists {
		s.client.HDel("voters:"+a.GetID(), u.GetID())
		return
	}
	s.client.HMSet("voters:"+a.GetID(), map[string]interface{}{u.GetID(): u.GetName()})
}
