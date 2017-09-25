package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EVENT string

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

type SendgridEvent []struct {
	Email       string   `json:"email"`
	Timestamp   int      `json:"timestamp"`
	SMTPID      string   `json:"smtp-id"`
	Event       string   `json:"event"`
	Category    []string `json:"category"`
	SgEventID   string   `json:"sg_event_id"`
	SgMessageID string   `json:"sg_message_id"`
	Response    string   `json:"response,omitempty"`
	Attempt     string   `json:"attempt,omitempty"`
	Useragent   string   `json:"useragent,omitempty"`
	IP          string   `json:"ip,omitempty"`
	URL         string   `json:"url,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	Status      string   `json:"status,omitempty"`
	AsmGroupID  int      `json:"asm_group_id,omitempty"`
}

func main() {
	router := gin.Default()
	router.POST("/v1/eventhook/", incomingWebhook)
	router.GET("/v1/mails", listMails)
	router.GET("/v1/mail/:uuid", mailStates)
	router.Run("0.0.0.0:3001")
}

func incomingWebhook(c *gin.Context) {
	//body, _ := ioutil.ReadAll(c.Request.Body)
	// fmt.Println("Headers:")
	// fmt.Println(c.Request.Header)
	//	fmt.Printf("%s", string(body))
	// fmt.Println("EVENT RECEIVED : ")
	var events SendgridEvent
	if err := c.BindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		fmt.Println(err.Error())
		return
	}
	for _, event := range events {
		fmt.Printf("%+v\n", event)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
	return
}

func listMails(c *gin.Context) {

}

func mailStates(c *gin.Context) {

}
