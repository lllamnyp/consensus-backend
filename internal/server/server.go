package server

import (
	"encoding/json"
	"errors"
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
var l = config.Logger

func verifyToken(r *http.Request) bool {
	tokenCookie, err := r.Cookie("AuthToken")
	tokenString := tokenCookie.Value
	if err != nil {
		return false
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			l.Error("Unexpected signing method", zap.Any("method", token.Header["alg"]))
			return nil, errors.New("Unexpected signing method: " + fmt.Sprintf("%v", token.Header["alg"]))
		}
		return hmacSampleSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		nafClaim, ok := claims["naf"]
		if !ok {
			l.Error("JWT token missing naf claim")
			return false
		}
		nafString, ok := nafClaim.(string)
		if !ok {
			l.Error("naf claim in token holds unexpected value", zap.Any("claim", nafClaim))
		}
		if naf, err := strconv.Atoi(nafString); err != nil || time.Now().Unix()*1000 > int64(naf) {
			return false
		}
		return true
	} else {
		fmt.Println(err)
		return false
	}

}

func Serve(p poll.Poll) {
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
