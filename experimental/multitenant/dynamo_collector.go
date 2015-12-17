package multitenant

import (
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

//https://github.com/aws/aws-sdk-go/wiki/common-examples

type dynamoDBCollector struct {
}

// NewDynamoDBCollector the reaper of souls
func NewDynamoDBCollector(url string) app.Collector {
	return &dynamoDBCollector{}
}

func (c *dynamoDBCollector) Report(context.Context) report.Report {
	return report.Report{}
}

func (c *dynamoDBCollector) WaitOn(context.Context, chan struct{}) {}

func (c *dynamoDBCollector) UnWait(context.Context, chan struct{}) {}

func (c *dynamoDBCollector) Add(context.Context, report.Report) {}
