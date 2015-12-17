package multitenant

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

//https://github.com/aws/aws-sdk-go/wiki/common-examples

const (
	tableName   = "reports"
	hourField   = "hour"
	tsField     = "ts"
	reportField = "report"
)

type dynamoDBCollector struct {
	db *dynamodb.DynamoDB
}

// NewDynamoDBCollector the reaper of souls
func NewDynamoDBCollector(url string) app.Collector {
	return &dynamoDBCollector{
		db: dynamodb.New(session.New(), aws.NewConfig()),
	}
}

func (c *dynamoDBCollector) createTable() error {
	params := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(hourField),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String(tsField),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String(reportField),
				AttributeType: aws.String("B"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(hourField),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(tsField),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := c.db.CreateTable(params)
	return err
}

func (c *dynamoDBCollector) Report(context.Context) report.Report {
	return report.Report{}
}

func (c *dynamoDBCollector) Add(_ context.Context, rep report.Report) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(rep); err != nil {
		panic(err)
	}

	now := time.Now()
	_, err := c.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			hourField: {
				N: aws.String(strconv.FormatInt(now.Unix()/3600, 10)),
			},
			tsField: {
				N: aws.String(strconv.FormatInt(now.UnixNano(), 10)),
			},
			reportField: {
				B: buf.Bytes(),
			},
		},
	})
	if err != nil {
		panic(err)
	}
}

func (c *dynamoDBCollector) WaitOn(context.Context, chan struct{}) {}

func (c *dynamoDBCollector) UnWait(context.Context, chan struct{}) {}
