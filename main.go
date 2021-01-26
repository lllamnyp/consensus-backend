/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/go-redis/redis"
	"github.com/lllamnyp/consensus-backend/internal/redisstore"
	"github.com/lllamnyp/consensus-backend/internal/server"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.String("redis-endpoint", "", "Specify the host and port of your redis backend")
	viper.BindPFlags(pflag.CommandLine)
	var s poll.Store
	if ep := viper.GetString("redis-endpoint"); ep == "" {
		s = poll.NewInMemoryStore()
	} else {
		rclient := redis.NewClient(&redis.Options{Addr: ep})
		s = redisstore.NewRedisStore(rclient)
	}
	p := poll.New(s)
	server.Serve(p)
}
