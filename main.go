package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"

	pb "github.com/clicrdv/ms-grpc-stubs/followservice"
	"github.com/clicrdv/ms-sendgrid-webhook/es"
	"github.com/gin-gonic/gin"
)

type EVENT string

var (
	esClient *es.ElasticsearchClient
)

const (
	SENT        = EVENT("SENT")      // SENT BY CLICRDV MAIL MICROSERVICE
	PROCESSED   = EVENT("PROCESSED") // sendgrid received the mail and it prepare it to be delivered
	DROPPED     = EVENT("DROPPED")   // Email has been removed by sendgrid or by remote smtp
	DEFERRED    = EVENT("DEFERRED")  // refused for the moment by remote smtp but with no valable reason, so sendgrid will retry during 72H
	BOUNCE      = EVENT("BOUNCE")    // bad address or stuff like that, remoete smtp refuse to deliver
	DELIVERED   = EVENT("DELIVERED") // email given to remote smtp with ok return
	OPEN        = EVENT("OPEN")      // email opened by recipient
	CLICK       = EVENT("CLICK")     //
	SPAM_REPORT = EVENT("SPAM REPORT")
	UNSUBSCRIBE = EVENT("UNSUBSCRIBE")
)

type SendgridEvent struct {
	Email       string    `json:"email"`
	Timestamp   int64     `json:"timestamp"`
	Date        time.Time `json:"@timestamp"`
	SMTPID      string    `json:"smtp-id"`
	Event       string    `json:"event"`
	Category    []string  `json:"category"` // trouver un moyen de gerer une ou n category
	SgEventID   string    `json:"sg_event_id"`
	SgMessageID string    `json:"sg_message_id"`
	Response    string    `json:"response,omitempty"`
	Attempt     string    `json:"attempt,omitempty"`
	Useragent   string    `json:"useragent,omitempty"`
	IP          string    `json:"ip,omitempty"`
	URL         string    `json:"url,omitempty"`
	Reason      string    `json:"reason,omitempty"`
	Status      string    `json:"status,omitempty"`
	AsmGroupID  int       `json:"asm_group_id,omitempty"`
	ClicRdvId   string    `json:"clicrdvid"'`
	Groupid     string
}

type GrpcServer struct{}

func (s *GrpcServer) NotifySentMail(ctx context.Context, followMail *pb.ClicRdvFollowMail) (*pb.SendMailStatus, error) {
	log.Print("Received GRPC Call")
	log.Printf("Received arguments : %s, %s", followMail.GetEmail(), followMail.GetUuid())
	categories := []string{"MS-MAIL"}
	sgEv := SendgridEvent{
		Category:  categories,
		Email:     followMail.GetEmail(),
		Event:     followMail.GetEvent(),
		ClicRdvId: followMail.GetUuid(),
		Groupid:   followMail.GetGroupId(),
		Timestamp: time.Now().Unix(),
	}

	StoreToEs(esClient, &sgEv)
	return &pb.SendMailStatus{}, nil
}

func StoreToEs(esClient *es.ElasticsearchClient, event *SendgridEvent) {
	event.Date = time.Unix(event.Timestamp, 0)
	log.Println("Timestamp is :", event.Timestamp)
	log.Println("Converted Timestamp is :", time.Unix(event.Timestamp, 0))
	esClient.StoreJson(event)
}

func main() {
	esClient = es.NewElasticsearchClient(os.Getenv("ES_URL"), "mail")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		router := gin.Default()
		router.POST("/v1/eventhook/", incomingWebhook(esClient))
		router.GET("/v1/mails", listMails)
		router.GET("/v1/mail/:uuid", mailStates)
		router.Run("0.0.0.0:3001")
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		log.Print("Starting microservice grpc listening on 50053")
		lis, err := net.Listen("tcp", "0.0.0.0:50053")
		if err != nil {
			log.Fatalf("Failed to listen : %v", err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterClicRdvFollowMailServiceServer(grpcServer, &GrpcServer{})
		grpcServer.Serve(lis)
		wg.Done()
	}()
	wg.Wait()
}

func incomingWebhook(esClient *es.ElasticsearchClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		//body, _ := ioutil.ReadAll(c.Request.Body)
		// fmt.Println("Headers:")
		// fmt.Println(c.Request.Header)
		//	fmt.Printf("%s", string(body))
		// fmt.Println("EVENT RECEIVED : ")
		var events []SendgridEvent
		if err := c.BindJSON(&events); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "json decoding : " + err.Error(),
				"status": http.StatusBadRequest,
			})
			log.Println(err.Error())
			return
		}
		for _, event := range events {
			StoreToEs(esClient, &event)
			log.Printf("%+v\n", event)
			log.Println("email is : ", event.Email)
			log.Println("Event is  : ", event.Event)
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}
}

func listMails(c *gin.Context) {

}

func mailStates(c *gin.Context) {

}
