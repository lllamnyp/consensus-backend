package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
)

var hmacSampleSecret = []byte(os.Getenv("SIGNING_SECRET"))

func verifyToken(r *http.Request) (string, bool) {
	tokenCookie, err := r.Cookie("AuthToken")
	tokenString := tokenCookie.Value
	if err != nil {
		return "", false
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %s: %v\n", "method", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		nafClaim, ok := claims["naf"]
		if !ok {
			fmt.Printf("JWT token missing naf claim\n")
			return "", false
		}
		nafString, ok := nafClaim.(string)
		if !ok {
			fmt.Printf("naf claim in token holds unexpected value %s: %v\n", "claim", nafClaim)
			return "", false
		}
		if naf, err := strconv.Atoi(nafString); err != nil || time.Now().Unix()*1000 > int64(naf) {
			return "", false
		}
		usrClaim, ok := claims["usr"]
		if !ok {
			fmt.Printf("JWT token missing usr claim\n")
			return "", false
		}
		usrString, ok := usrClaim.(string)
		if !ok {
			fmt.Printf("usr claim in token holds unexpected value %s: %v\n", "claim", usrClaim)
			return "", false
		}

		return usrString, true
	} else {
		fmt.Println(err)
		return "", false
	}
}

func Serve(p poll.Poll) {
	listAnswers := func(w http.ResponseWriter, r *http.Request) {
		var user string
		var ok bool
		if user, ok = verifyToken(r); !ok {
			w.Write([]byte("{}"))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		if r.Method == "GET" {
			answers := p.ListAnswers()
			u := poll.NewUser(user)
			for id := range answers {
				answers[id].WithUser(u)
			}
			answersJSON, err := json.Marshal(answers)
			if err != nil {
				fmt.Printf("Failed to marshal answer list\n%+v\n", answers)
				return
			}
			w.Write(answersJSON)
			fmt.Printf("Serving answer list... %s: %v\n", "Length", len(answers))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Bad request for answer list\n")
		}
	}
	addAnswer := func(w http.ResponseWriter, r *http.Request) {
		var user string
		var ok bool
		if user, ok = verifyToken(r); !ok {
			return
		}
		u := poll.NewUser(user)
		if r.Method == "POST" {
			r.ParseForm()
			content := r.Form["content"][0]
			p.AddAnswer(u, poll.NewAnswer(content))
			fmt.Printf("Posting answer %s: %v\n", "Value", content)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Bad request to submit answer\n")
		}
	}
	toggleVote := func(w http.ResponseWriter, r *http.Request) {
		var user string
		var ok bool
		if user, ok = verifyToken(r); !ok {
			return
		}
		u := poll.NewUser(user)
		paths := strings.Split(r.URL.Path, "/")
		if len(paths) < 3 {
			fmt.Printf("Unexpected vote request URI: %+v\n", r.URL.Path)
			return
		}
		id := paths[len(paths)-1]
		a, err := p.GetAnswerByID(id)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		p.ToggleVote(u, a)
	}
	http.HandleFunc("/api/list", listAnswers)
	http.HandleFunc("/api/add", addAnswer)
	http.HandleFunc("/api/vote/", toggleVote)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}
