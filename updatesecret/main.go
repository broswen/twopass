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
	"golang.org/x/crypto/bcrypt"
)

type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

type UpdateSecretRequest struct {
	Id        string `json:"id" validate:"required"`
	NewSecret string `json:"newSecret" validate:"required"`
	PassOne   string `json:"passOne" validate:"required"`
	PassTwo   string `json:"passTwo" validate:"required"`
}

type UpdateSecretResponse struct {
	Id string `json:"id" validate:"required"`
}

var validate *validator.Validate

var ddbClient *dynamodb.Client

func Handler(ctx context.Context, request Request) (Response, error) {

	input := &UpdateSecretRequest{}
	err := json.Unmarshal([]byte(request.Body), input)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	err = validate.Struct(input)
	if err != nil {
		return Response{StatusCode: 400, Body: err.Error()}, nil
	}

	getItemParams := &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("SECRETSTABLE")),
		Key:       map[string]types.AttributeValue{"PK": &types.AttributeValueMemberS{Value: input.Id}},
	}

	getResponse, err := ddbClient.GetItem(context.TODO(), getItemParams)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	if getResponse.Item == nil {
		return Response{StatusCode: 404}, nil
	}

	//check passone hash matches
	if err = bcrypt.CompareHashAndPassword([]byte(getResponse.Item["passOne"].(*types.AttributeValueMemberS).Value), []byte(input.PassOne)); err != nil {
		return Response{StatusCode: 401}, nil
	}
	//check passtwo hash matches
	if err = bcrypt.CompareHashAndPassword([]byte(getResponse.Item["passTwo"].(*types.AttributeValueMemberS).Value), []byte(input.PassTwo)); err != nil {
		return Response{StatusCode: 401}, nil
	}

	updateItemInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv("SECRETSTABLE")),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: input.Id},
		},
		UpdateExpression: aws.String("SET #s = :s"),
		ExpressionAttributeNames: map[string]string{
			"#s": "secret",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberS{Value: input.NewSecret},
		},
	}

	_, err = ddbClient.UpdateItem(context.TODO(), updateItemInput)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	response := &UpdateSecretResponse{input.Id}

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
