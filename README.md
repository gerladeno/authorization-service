# authorization-service

## methods

#### /v1/authenticate
```shell
curl "http://0.0.0.0:3000/public/v1/authenticate?phone=+79260806722"
```

```json
{"data":"Ok"}
```

#### /v1/signIn
```shell
curl "http://0.0.0.0:3000/public/v1/signIn?phone=+79260806722&code=8726"
```

```json
{"data":{"token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IiIsInVzZXJuYW1lIjoiIiwicGhvbmUiOiIgNzkyNjA4MDY3MjIifQ.ZkNFVNpNYamM2p_KhMSv0Uy2Xk7tbrwq-NhsTCKWFYouWHAst4uU9u1z3jGxCxrbHUX4Y-mqOGbhyERFZQbF7PxHBD63jCVW8hiCmAkMP4yZVVBopBHPEKk2oQhUCW_85F6W6L2aW6VO-f1ZEB7hkp9y4xYAtiCWzfynjBs04YIFVesvwYY10uGCRczhIXLUSQam7fi3-jxBbRr6RB4i2rnl-mMrnOXjInDt50v08_-nlvhg8XQqJ-kA06oHUxr5nQHakI1B2YkaN93KaZP8I_MjDP-yBpbUQPTLejECSYSnHzfoxvdn6S3AIXCt6KtD4IzJx9WkxM-grDQ3bxhxfg"}}
```

#### /v1/verify

```shell
curl "http://0.0.0.0:3000/public/v1/verify?token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IiIsInVzZXJuYW1lIjoiIiwicGhvbmUiOiIgNzkyNjA4MDY3MjIifQ.ZkNFVNpNYamM2p_KhMSv0Uy2Xk7tbrwq-NhsTCKWFYouWHAst4uU9u1z3jGxCxrbHUX4Y-mqOGbhyERFZQbF7PxHBD63jCVW8hiCmAkMP4yZVVBopBHPEKk2oQhUCW_85F6W6L2aW6VO-f1ZEB7hkp9y4xYAtiCWzfynjBs04YIFVesvwYY10uGCRczhIXLUSQam7fi3-jxBbRr6RB4i2rnl-mMrnOXjInDt50v08_-nlvhg8XQqJ-kA06oHUxr5nQHakI1B2YkaN93KaZP8I_MjDP-yBpbUQPTLejECSYSnHzfoxvdn6S3AIXCt6KtD4IzJx9WkxM-grDQ3bxhxfg"
```

```json
{"data":{"userId":""}}
```
id will be provided if user is found in DB