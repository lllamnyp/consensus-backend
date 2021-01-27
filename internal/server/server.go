package server

import (
	"encoding/json"
	"net/http"

	"github.com/lllamnyp/consensus-backend/internal/config"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
	"go.uber.org/zap"
)

func Serve(p poll.Poll) {
	l := config.Logger
	listAnswers := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if r.Method == "GET" {
			answers, _ := json.Marshal(p.ListAnswers())
			w.Write(answers)
			l.Info("Serving answer list...", zap.Int("Length", len(answers)))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			l.Error("Bad request for answer list")
		}
	}
	addAnswer := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			content := r.Form["content"][0]
			p.AddAnswer(nil, poll.NewAnswer(content))
			l.Info("Posting answer", zap.String("Value", content))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			l.Error("Bad request to submit answer")
		}
	}
	http.HandleFunc("/api/list", listAnswers)
	http.HandleFunc("/api/add", addAnswer)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}
