package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KongZ/phabrick/internal/conduit"
	"github.com/KongZ/phabrick/internal/config"
	slack "github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/colornames"
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
		log.Debugf("%+v", webhookReq)
		for _, objectType := range webhook.Config.Channels.ObjectTypes {
			if webhookReq.Object.Type == objectType {
				var task conduit.ManiphestResponse
				var assignee conduit.UserResponse
				var project conduit.ProjectResponse
				var transactions []conduit.TransactionResponse
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
							if len(webhookReq.Transactions) > 0 {
								wg.Add(1)
								go func() {
									defer wg.Done()
									taskTransactions, err := con.GetTransactions([]string{task.ID})
									if err == nil {
										transactions = taskTransactions[task.ID]
										for _, tran := range transactions {
											if tran.TransactionPHID == webhookReq.Transactions[0].Phid {
												// actor
												wg.Add(1)
												go func() {
													defer wg.Done()
													users, err := con.QueryUser([]string{tran.AuthorPHID})
													if err == nil {
														actor = users[0]
													} else {
														log.Errorf("Error while quering PHID %s, %v", tran.AuthorPHID, err)
													}
												}()
												break
											}
										}
									} else {
										log.Errorf("Error while quering maniphest ID %v, %v", maniphests[0].ID, err)
									}
								}()
							}
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
				fields := []slack.AttachmentField{}
				if webhook.Config.Slack.ShowAssignee {
					fields = append(fields, slack.AttachmentField{
						Title: "Assignee",
						Value: assignee.UserName,
						Short: true,
					})
				}
				// if webhook.Config.Slack.ShowAuthor {
				// 	users, err := con.QueryUser([]string{task.AuthorPHID})
				// 	if err == nil {
				// 		fields = append(fields, slack.AttachmentField{
				// 			Title: "Author",
				// 			Value: users[0].UserName,
				// 			Short: true,
				// 		})
				// 	} else {
				// 		log.Errorf("Error while quering PHID %s, %v", task.AuthorPHID, err)
				// 	}
				// }
				var actions []string
				if transactions != nil {
					for _, transaction := range transactions {
						if transaction.DateCreated == task.DateModified {
							switch transaction.TransactionType {
							case "status":
								actions = append(actions, fmt.Sprintf("has been `%s`", transaction.NewValue.Description))
								break
							case "core:comment":
								actions = append(actions, fmt.Sprintf("added a comment"))
								fields = append(fields, slack.AttachmentField{
									Title: "Comment",
									Value: transaction.Comments,
									Short: false,
								})
								break
							case "description":
								fields = append(fields, slack.AttachmentField{
									Title: "Description",
									Value: transaction.NewValue.Description,
									Short: false,
								})
								break
							case "core:subscribers":
								subscribers, e := con.QueryUser(transaction.NewValue.Users)
								if e == nil && len(subscribers) > 0 {
									actions = append(actions, fmt.Sprintf("%v was added to subscribers", subscribers[0].UserName))
								}
								break
							case "reassign":
								actions = append(actions, fmt.Sprintf("task was assigned to %s", assignee.UserName))
								break
							case "core:create":
								actions = append(actions, fmt.Sprintf("was created"))
								break
							case "core:columns":
								newValueColumn := transaction.NewValue.Column[0]
								if len(newValueColumn.FromColumnPHIDs) > 0 {
									var fromColumn string
									for k := range newValueColumn.FromColumnPHIDs {
										fromColumn = k
										break
									}
									columns, e := con.QueryColumn([]string{fromColumn, newValueColumn.ColumnPHID})
									if e == nil && len(columns) == 2 {
										actions = append(actions, fmt.Sprintf("moved from %s to %s ", columns[0].Name, columns[1].Name))
									}
								}
								break
							}
						}
					}
				}
				if len(actions) > 0 {
					slackAPI := slack.New(webhook.Config.Slack.Token)
					channelID, ok := webhook.Config.Channels.Projects[project.ID]
					if !ok {
						channelID = webhook.Config.Channels.Projects["default"]
					}
					if channelID == "" {
						log.Infof("[%s] No channel found for project %s", task.ID, project.ID)
						return
					}
					log.Debugf("Sending message to %s", channelID)
					attachment := slack.Attachment{
						Title:      fmt.Sprintf("[%s] %s", task.ObjectName, task.Title),
						TitleLink:  task.URI,
						Text:       strings.Join(actions, ", "),
						Fields:     fields,
						Footer:     fmt.Sprintf("<%s/project/profile/%s|on %s>", webhook.Config.Phabricator.URL, project.ID, project.Name),
						Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
						FooterIcon: "https://raw.githubusercontent.com/phacility/phabricator/master/webroot/rsrc/favicons/favicon-16x16.png",
					}
					if webhook.Config.Slack.ShowAuthor {
						attachment.AuthorName = actor.UserName
						attachment.AuthorIcon = actor.Image
					}
					cn, found := colornames.Map[strings.ToLower(task.PriorityColor)]
					if found {
						attachment.Color = fmt.Sprintf("#%02x%02x%02x", cn.R, cn.G, cn.B)
					}
					// it is not possible to get icon outside Phabricator. `file.info` return a link to HTML
					// if project.Icon != "" {
					// 	attachment.FooterIcon = project.Icon
					// }
					channelID, timestamp, err := slackAPI.PostMessage(channelID,
						slack.MsgOptionText("", false),
						slack.MsgOptionUsername(webhook.Config.Slack.Username),
						slack.MsgOptionTS(fmt.Sprintf("%d", time.Now().Unix())),
						slack.MsgOptionAttachments(attachment))
					if err != nil {
						log.Errorf("%s\n", err)
						return
					}
					log.Infof("Message successfully sent to channel %s at %s", channelID, timestamp)
				} else {
					log.Infoln("No actions are required")
				}
			}
		}
	} else {
		log.Errorf("No Slack token provided")
	}
}
