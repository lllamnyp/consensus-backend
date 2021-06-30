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
	"github.com/lllamnyp/consensus-backend/internal/speakers"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
)

var hmacSampleSecret = []byte(os.Getenv("SIGNING_SECRET"))

func verifyToken(r *http.Request) (string, string, bool) {
	tokenCookie, err := r.Cookie("AuthToken")
	tokenString := tokenCookie.Value
	if err != nil {
		return "", "", false
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
			return "", "", false
		}
		nafString, ok := nafClaim.(string)
		if !ok {
			fmt.Printf("naf claim in token holds unexpected value %s: %v\n", "claim", nafClaim)
			return "", "", false
		}
		if naf, err := strconv.Atoi(nafString); err != nil || time.Now().Unix()*1000 > int64(naf) {
			return "", "", false
		}
		usrClaim, ok := claims["username"]
		if !ok {
			fmt.Printf("JWT token missing usr claim\n")
			return "", "", false
		}
		usrString, ok := usrClaim.(string)
		if !ok {
			fmt.Printf("usr claim in token holds unexpected value %s: %v\n", "claim", usrClaim)
			return "", "", false
		}
		subClaim, ok := claims["sub"]
		if !ok {
			fmt.Printf("JWT token missing usr claim\n")
			return "", "", false
		}
		subString, ok := subClaim.(string)
		if !ok {
			fmt.Printf("sub claim in token holds unexpected value %s: %v\n", "claim", subClaim)
			return "", "", false
		}
		return subString, usrString, true
	} else {
		fmt.Println(err)
		return "", "", false
	}
}

func Serve(p poll.Poll) {
	listAnswers := func(w http.ResponseWriter, r *http.Request) {
		var login string
		var user string
		var ok bool
		if login, user, ok = verifyToken(r); !ok {
			w.Write([]byte("{}"))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		if r.Method == "GET" {
			answers := p.ListAnswers()
			u := poll.NewUser(login, user)
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
		var login string
		var user string
		var ok bool
		if login, user, ok = verifyToken(r); !ok {
			return
		}
		u := poll.NewUser(login, user)
		if r.Method == "POST" {
			r.ParseForm()
			anonymous := r.Form["anonymous"][0] == "true"
			content := r.Form["content"][0]
			addressee, err := strconv.Atoi(r.Form["addressee"][0])
			if err != nil {
				addressee = 0
			}
			uAddressee := poll.NewUser(speakers.ReverseLookup(addressee))
			p.AddAnswer(u, poll.NewAnswer(content, u, uAddressee, anonymous))
			fmt.Printf("Posting answer %s: %v\n", "Value", content)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Bad request to submit answer\n")
		}
	}
	toggleVote := func(w http.ResponseWriter, r *http.Request) {
		var login string
		var user string
		var ok bool
		if login, user, ok = verifyToken(r); !ok {
			return
		}
		u := poll.NewUser(login, user)
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
	respond := func(w http.ResponseWriter, r *http.Request) {
		var login string
		var user string
		var ok bool
		if login, user, ok = verifyToken(r); !ok {
			return
		}
		u := poll.NewUser(login, user)
		paths := strings.Split(r.URL.Path, "/")
		if len(paths) < 3 {
			fmt.Printf("Unexpected response URI: %+v\n", r.URL.Path)
			return
		}
		id := paths[len(paths)-1]
		a, err := p.GetAnswerByID(id)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		if r.Method == "POST" {
			r.ParseForm()
			response := r.Form["response"][0]
			a.WithResponse(response)
			p.Respond(u, a)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Bad request to submit response\n")
		}
	}
	http.HandleFunc("/api/list", listAnswers)
	http.HandleFunc("/api/add", addAnswer)
	http.HandleFunc("/api/vote/", toggleVote)
	http.HandleFunc("/api/respond/", respond)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}
