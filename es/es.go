package es

import (
	"fmt"

	"golang.org/x/net/context"

	elastic "gopkg.in/olivere/elastic.v5"

	"time"
)

const (
	//ElasticsearchType designate the type put in messages
	ElasticsearchType     = "sendgrid_mail"
	logstashIndexWildcard = "logstash-*"
)

//ElasticsearchClient represent an ElasticsearchClient
type ElasticsearchClient struct {
	Client  *elastic.Client
	ESIndex string
}

func MailIndex() string {
	now := time.Now()
	return fmt.Sprintf("mail-%d.%.2d.%.2d", now.Year(), now.Month(), now.Day())
}

//NewElasticsearchClient allow to create an ElasticsearchClient
func NewElasticsearchClient(URL string, index string) *ElasticsearchClient {
	// Create a client
	client, err := elastic.NewClient(
		elastic.SetMaxRetries(10),
		elastic.SetSniff(false),
		elastic.SetURL(URL),
		elastic.SetBasicAuth("elastic", "changeme"),
	)

	fmt.Println("Sending message to ES : ", URL)

	if err != nil {
		fmt.Println("Error while initializing Elasticsearch client : ", err.Error())
	}
	client.CreateIndex(index).Do(context.TODO())
	return &ElasticsearchClient{Client: client, ESIndex: index}
}

//ForwardMessage is used to forward a message in elasticsearch
func (ESClient *ElasticsearchClient) StoreJson(m interface{}) {
	// Create an index
	if ESClient.ESIndex == "mail" {
		ESClient.ESIndex = MailIndex()
	}

	fmt.Println("Sending to Index : ", ESClient.ESIndex)
	ESClient.Client.Index().
		Index(ESClient.ESIndex).
		Type(ElasticsearchType).
		BodyJson(m).
		Do(context.TODO())
	fmt.Println("Sent")
}
