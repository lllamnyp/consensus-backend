package redisstore

import (
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
	//s.client.("votes", map[string]interface{}{a.GetID(): a.GetContent()})
	return
}

func (s *redisstore) ListAnswers() map[string]poll.Answer {
	ret := make(map[string]poll.Answer)
	val := s.client.HGetAll("answers").Val()
	for k, v := range val {
		ret[k] = poll.NewAnswer(v)
	}
	return ret
}
