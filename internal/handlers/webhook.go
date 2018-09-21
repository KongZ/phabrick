package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/KongZ/phabrick/internal/conduit"
	"github.com/KongZ/phabrick/internal/config"
	slack "github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

// WebhookRequest a request object from Phabricator
type WebhookRequest struct {
	Object struct {
		Type string `json:"type"`
		Phid string `json:"phid"`
	} `json:"object"`
	Triggers []struct {
		Phid string `json:"phid"`
	} `json:"triggers"`
	Action struct {
		Test   bool `json:"test"`
		Silent bool `json:"silent"`
		Secure bool `json:"secure"`
		Epoch  int  `json:"epoch"`
	} `json:"action"`
	Transactions []struct {
		Phid string `json:"phid"`
	} `json:"transactions"`
}

// Webhook new Webhook handler
type Webhook struct {
	Config *config.Config
}

// content returns a simple HTTP handler function which writes a text response.
func (webhook *Webhook) receiveNotify(w http.ResponseWriter, r *http.Request) {
	if webhook.Config.Slack.Token != "" {
		con := &conduit.Conduit{Config: webhook.Config}
		var webhookReq WebhookRequest
		defer r.Body.Close()

		// requestDump, e := httputil.DumpRequest(r, true)
		// if e != nil {
		// 	fmt.Println(e)
		// }
		// fmt.Println(string(requestDump))

		err := json.NewDecoder(r.Body).Decode(&webhookReq)
		if err != nil {
			log.Errorf("Invalid request from webhook, %v", err)
		}
		log.Printf("Received %v", webhookReq.Object)
		for _, objectType := range webhook.Config.Channels.ObjectTypes {
			if webhookReq.Object.Type == objectType {
				var task conduit.ManiphestResponse
				var assignee conduit.UserResponse
				var project conduit.ProjectResponse
				var transaction conduit.TransactionResponse
				var actor conduit.UserResponse
				var wg sync.WaitGroup
				wg.Add(1)

				if webhookReq.Object.Phid != "" {
					go func() {
						defer wg.Done()
						// maniphest
						maniphests, err := con.QueryManiphest([]string{webhookReq.Object.Phid})
						if err != nil {
							log.Errorf("Error while quering PHID %s, %v", webhookReq.Object.Phid, err)
							return
						}
						if len(maniphests) > 0 {
							task = maniphests[0]
							log.Printf("%#v", task)
							// project
							wg.Add(1)
							go func() {
								defer wg.Done()
								projects, err := con.QueryProject(maniphests[0].ProjectPHIDs)
								if err == nil {
									project = projects[0]
								} else {
									log.Errorf("Error while quering PHID %v, %v", maniphests[0].ProjectPHIDs, err)
								}
							}()
							// transactions
							wg.Add(1)
							go func() {
								defer wg.Done()
								transactions, err := con.GetTransactions([]string{maniphests[0].ID})
								if err == nil {
									for _, tran := range transactions {
										if tran.TransactionPHID == webhookReq.Transactions[0].Phid {
											transaction = tran
											// actor
											wg.Add(1)
											go func() {
												defer wg.Done()
												users, err := con.QueryUser([]string{transaction.AuthorPHID})
												if err == nil {
													actor = users[0]
												} else {
													log.Errorf("Error while quering PHID %s, %v", transaction.AuthorPHID, err)
												}
											}()
											break
										}
									}
								} else {
									log.Errorf("Error while quering maniphest ID %v, %v", maniphests[0].ID, err)
								}
							}()
							// assignee
							wg.Add(1)
							go func() {
								defer wg.Done()
								users, err := con.QueryUser([]string{maniphests[0].OwnerPHID})
								if err == nil {
									assignee = users[0]
								} else {
									log.Errorf("Error while quering PHID %s, %v", maniphests[0].OwnerPHID, err)
								}
							}()
						}
					}()
				}
				// merge
				wg.Wait()
				slackAPI := slack.New(webhook.Config.Slack.Token)
				channelID, ok := webhook.Config.Channels.Projects[project.ID]
				if !ok {
					channelID = webhook.Config.Channels.Projects["default"]
				}
				if channelID == "" {
					log.Infof("No channel found for project %s", project.ID)
					return
				}
				log.Infof("Sending message to %s", channelID)
				fields := []slack.AttachmentField{}
				if webhook.Config.Slack.ShowAssignee {
					fields = append(fields, slack.AttachmentField{
						Title: "Assignee",
						Value: assignee.UserName,
						Short: true,
					})
				}
				if webhook.Config.Slack.ShowAuthor {
					users, err := con.QueryUser([]string{task.AuthorPHID})
					if err == nil {
						fields = append(fields, slack.AttachmentField{
							Title: "Author",
							Value: users[0].UserName,
							Short: true,
						})
					} else {
						log.Errorf("Error while quering PHID %s, %v", task.AuthorPHID, err)
					}
				}
				var action string
				switch transaction.TransactionType {
				case "status":
					action = fmt.Sprintf("has been %s by %s", transaction.NewValue, actor.UserName)
					break
				case "core:comment":
					action = fmt.Sprintf("%s add a comment", actor.UserName)
					fields = append(fields, slack.AttachmentField{
						Title: "Comment",
						Value: transaction.Comments,
						Short: false,
					})
					break
				case "description":
					action = fmt.Sprintf("%s update description", actor.UserName)
					fields = append(fields, slack.AttachmentField{
						Title: "Description",
						Value: transaction.NewValue,
						Short: false,
					})
					break
				case "core:subscribers":
					action = fmt.Sprintf("%s add subscribers", actor.UserName)
					break
				case "reassign":
					action = fmt.Sprintf("%s assign task to %s", actor.UserName, assignee.UserName)
					break
				case "core:create":
					action = fmt.Sprintf("was created by %s", actor.UserName)
					break
				}
				params := slack.PostMessageParameters{}
				attachment := slack.Attachment{
					AuthorName: task.Title,
					AuthorLink: task.URI,
					Text:       action,
					Fields:     fields,
					Footer:     project.Name,
					FooterIcon: "https://raw.githubusercontent.com/phacility/phabricator/master/webroot/rsrc/favicons/favicon-16x16.png",
				}
				params.Attachments = []slack.Attachment{attachment}
				params.Username = webhook.Config.Slack.Username
				params.ThreadTimestamp = fmt.Sprintf("%d", time.Now().Unix())
				channelID, timestamp, err := slackAPI.PostMessage(channelID, task.Title, params)
				if err != nil {
					log.Errorf("%s\n", err)
					return
				}
				log.Infof("Message successfully sent to channel %s at %s", channelID, timestamp)
			}
		}
	} else {
		log.Errorf("No Slack token provided")
	}
}
