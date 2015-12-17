package multitenant

import (
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

func (c *dynamoDBCollector) Report() report.Report {
	return report.Report{}
}

func (c *dynamoDBCollector) WaitOn(chan struct{}) {}

func (c *dynamoDBCollector) UnWait(chan struct{}) {}

func (c *dynamoDBCollector) Add(report.Report) {}
