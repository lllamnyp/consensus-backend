package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lllamnyp/consensus-backend/internal/config"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
	"go.uber.org/zap"
)

var hmacSampleSecret = os.Getenv("SIGNING_SECRET")

func verifyToken(r *http.Request) bool {
	tokenCookie, err := r.Cookie("AuthToken")
	tokenString := tokenCookie.Value
	if err != nil {
		return false
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if naf, err := strconv.Atoi(claims["naf"].(string)); err != nil || time.Now().Unix()*1000 > int64(naf) {
			return false
		}
		return true
	} else {
		fmt.Println(err)
		return false
	}

}

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
		if !verifyToken(r) {
			return
		}
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
