# consensus-backend

---

backend for online polls with user-definable answers

## API

```
Authentication: Request headers
X-Auth-Hash: <value of 'hash´ cookie>
X-Auth-Log: <value of 'log´ cookie>

ListAnswers
GET /api/list
Request headers:
  X-Auth-Hash: <value of 'hash´ cookie>
  X-Auth-Log: <value of 'log´ cookie>
Response headers:
  Content-Type: application/json
Response body:
  []answer
Description: get JSON dump of answers

AddAnswer
POST /api/add
FormData
  // number of addressee of the question; 0 is null
  addressee int
  // should the asker's identity be hidden?
  anonymous bool
  // question body
  content string
Request headers:
  Content-Type: application/x-www-form-urlencoded
  X-Auth-Hash: <value of 'hash´ cookie>
  X-Auth-Log: <value of 'log´ cookie>
Description:
  Add a new answer to the list

Toggle Vote
GET /api/vote/{answerID}
Request headers:
  X-Auth-Hash: <value of 'hash´ cookie>
  X-Auth-Log: <value of 'log´ cookie>

Respond
POST /api/respond/{answerID}
Request headers:
  Content-Type: application/x-www-form-urlencoded
  X-Auth-Hash: <value of 'hash´ cookie>
  X-Auth-Log: <value of 'log´ cookie>

---

Models:

answer:
  {
    // id is the sha1 sum of the content encoded with base64
    id string
    // who asked the question?
    asker     string
    // to whom was it addressed?
    addressee int
    // question body
    content   string
    // response from addressee
    response  string
    // has the current user upvoted this question?
    upvoted   bool
    // votecount of question
    votes     int
    // unixtime when the question was added
    timestamp int
  }
```
