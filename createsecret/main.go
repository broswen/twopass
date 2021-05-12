package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

type CreateSecretRequest struct {
	Secret  string `json:"secret" validate:"required"`
	PassOne string `json:"passOne" validate:"required"`
	PassTwo string `json:"passTwo" validate:"required"`
}

type CreateSecretResponse struct {
	Id string `json:"id" validate:"required"`
}

var validate *validator.Validate

var ddbClient *dynamodb.Client

func Handler(ctx context.Context, request Request) (Response, error) {

	input := &CreateSecretRequest{}
	err := json.Unmarshal([]byte(request.Body), input)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	err = validate.Struct(input)
	if err != nil {
		return Response{StatusCode: 400, Body: err.Error()}, nil
	}

	id := ksuid.New()

	passOneHash, err := bcrypt.GenerateFromPassword([]byte(input.PassOne), bcrypt.DefaultCost)
	if err != nil {
		return Response{StatusCode: 500}, err
	}
	passTwoHash, err := bcrypt.GenerateFromPassword([]byte(input.PassTwo), bcrypt.DefaultCost)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	log.Println(string(passOneHash), string(passTwoHash), input.Secret)

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("SECRETSTABLE")),
		Item: map[string]types.AttributeValue{
			"PK":      &types.AttributeValueMemberS{Value: id.String()},
			"passOne": &types.AttributeValueMemberS{Value: string(passOneHash)},
			"passTwo": &types.AttributeValueMemberS{Value: string(passTwoHash)},
			"secret":  &types.AttributeValueMemberS{Value: input.Secret},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	_, err = ddbClient.PutItem(context.TODO(), putItemInput)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	response := &CreateSecretResponse{id.String()}

	var buf bytes.Buffer

	body, err := json.Marshal(response)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode: 200,
		Body:       buf.String(),
	}

	return resp, nil
}

func init() {
	validate = validator.New()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
}

func main() {
	lambda.Start(Handler)
}
