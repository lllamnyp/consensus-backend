# consensus-backend

---

backend for online polls with user-definable answers

## API

```
Authentication: Cookie
Name: AuthToken
Description: The AuthToken cookie holds a JWT token with a 'username´ claim. The signing method is HMAC.

ListAnswers
GET /api/list
Response headers:
  Content-Type: application/json
Response body:
{
  answerID: answer,
  ...
}
Description: get JSON dump of answers

AddAnswer
POST /api/add
FormData
  content string
Request headers:
  Content-Type: application/x-www-form-urlencoded
Description:
  Add a new answer to the list

Vote
GET /api/vote/{answerID}

---

Models:

answer:
  {
    content: string
    upvoted: bool
    votes: int
  }

answerID is the sha1 sum of its content encoded with base64
```
