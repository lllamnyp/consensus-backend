package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	// "strconv"
	"strings"
	// "time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lllamnyp/consensus-backend/internal/speakers"
	"github.com/lllamnyp/consensus-backend/pkg/poll"
)

type AddRequest struct {
	Content   string `json:"content"`
	Anonymous bool   `json:"anonymous"`
	Addressee int    `json:"addressee"`
}

type RespondRequest struct {
	Response string `json:"response"`
}

var hmacSampleSecret = []byte(os.Getenv("SIGNING_SECRET"))

func verifyToken(r *http.Request) (string, string, bool) {
	tokenHeader := r.Header["Authorization"][0]
	tokenStrings := strings.Split(tokenHeader, " ")
	tokenString := tokenStrings[len(tokenStrings)-1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %s: %v\n", "method", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})
	if err != nil {
		fmt.Printf("Failed parsing token, error: %s.\n", err)
	}

	/*
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
		}*/
	var id string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id = claims["data"].(map[string]interface{})["user"].(map[string]interface{})["id"].(string)
		fmt.Println(id)
	}
	reqURL, _ := url.Parse(os.Getenv("USER_ENDPOINT") + id)
	reqBody := ioutil.NopCloser(strings.NewReader(""))
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req := &http.Request{
		Method: "GET",
		URL:    reqURL,
		Header: map[string][]string{
			"Authorization": {"Bearer " + tokenString},
		},
		Body: reqBody,
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request, err: %s.\n", err)
		return "", "", false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body, err: %s.\n", err)
		return "", "", false
	}
	resp.Body.Close()
	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		fmt.Printf("Error unmarshaling body, err: %s.\n", err)
		return "", "", false
	}
	fmt.Println(bodyMap["user_email"], bodyMap["name"])
	return bodyMap["user_email"].(string), bodyMap["name"].(string), token.Valid
}

func getEmail(login string) string {
	dom := os.Getenv("RESTRICTED_DOMAIN")
	if len(dom) > 0 {
		parts := strings.Split(login, "@")
		return parts[0] + `@` + dom
	}
	return login
}

func Serve(p poll.Poll) {
	email := func(w http.ResponseWriter, r *http.Request) {
		var login string
		var ok bool
		if login, _, ok = verifyToken(r); !ok {
			w.Write([]byte("{}"))
			return
		}
		email := getEmail(login)
		w.Write([]byte(`{"email":"` + email + `"}`))
	}
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
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Printf("Could not read body. Error: %s.\n", err)
				return
			}
			var addRequest AddRequest
			err = json.Unmarshal(body, &addRequest)
			if err != nil {
				fmt.Printf("Could not parse body. Error: %s.\n", err)
				return
			}
			login, name := speakers.ReverseLookup(addRequest.Addressee)
			uAddressee := poll.NewUser(login, name)
			p.AddAnswer(u, poll.NewAnswer(addRequest.Content, u, uAddressee, addRequest.Anonymous))
			fmt.Printf("Posting answer %s: %v\n", "Value", addRequest.Content)
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
		if getEmail(a.GetAddressee().GetLogin()) != getEmail(login) {
			return
		}
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Printf("Could not read body. Error: %s.\n", err)
				return
			}
			var respondRequest RespondRequest
			err = json.Unmarshal(body, &respondRequest)
			if err != nil {
				fmt.Printf("Could not parse body. Error: %s.\n", err)
				return
			}
			a.WithResponse(respondRequest.Response)
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
	http.HandleFunc("/api/email", email)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}
