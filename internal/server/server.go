package server

import (
	"encoding/json"
	"net/http"

	"github.com/lllamnyp/consensus-backend/pkg/poll"
)

func Serve(p poll.Poll) {
	listAnswers := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if r.Method == "GET" {
			answers, _ := json.Marshal(p.ListAnswers())
			w.Write(answers)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	addAnswer := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			content := r.Form["content"][0]
			p.AddAnswer(nil, poll.NewAnswer(content))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	http.HandleFunc("/api/list", listAnswers)
	http.HandleFunc("/api/add", addAnswer)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}
