## TwoPass

TwoPass is a secret access manager API created with AWS API Gateway, DynamoDB, and AWS Lambda with Golang.

TwoPass stores secrets that can only be access if both passwords are presented. This allows two parties to share a single secret and prevent the other from accessing it without both parties present.

* Access passwords are hashed with bcrypt
* Secret is encrypted using both access passwords (TODO)
* API is secured with API keys and throttling

1. Create a secret
```json
{
    "secret": "this is secret",
    "passOne": "first password",
    "passTwo": "second password"
}
```
```json
{
    "id":"abc123"
}
```

2. Access the secret
```json
{
    "id": "abc123",
    "passOne": "first password",
    "passTwo": "second password"
}
```
```json
{
    "id":"abc123",
    "secret": "this is secret"
}
```

3. Update the secret
```json
{
    "id": "abc123",
    "newSecret": "new secret",
    "passOne": "first password",
    "passTwo": "second password"
}
```

4. Delete the secret
```json
{
    "id": "abc123",
    "passOne": "first password",
    "passTwo": "second password"
}
```

![diagram](./images/diagram.png)


### TODO
- [ ] Refactor request/response types into single package
- [ ] Refactor creation of DynamoDBClient
- [ ] Handle HTTP errors better
- [ ] Encrypt secret using access passwords