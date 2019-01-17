# Phabrick
A Slack Bot for Phabricator

This bot will use the latest API [Herald](https://secure.phabricator.com/herald/) to monitor events 
and send webhooks to the server. The Phabrick server will receive the requests from Herald then
use Conduit API to query responding detail and send notification to Slack

## Build
* Uses docker to build the image
```
docker build -t asia.gcr.io/kongz/phabrick:v1.3 -f deployments/Dockerfile .
```

## Installation

* Uses Helm Chart to install on Kubernetes
* Configure Slack Token, Phabricator URL and Token
* Configure mapping between Phabricator project ID and Slack channels
  * You can use `default` to send all unmatched project ID to this Slack channels
  

## phabrick.yaml configuration

Example

```yaml
    slack:
      token: xoxb-xxxxxxx
      username: wall-e
      showAssignee: false
      showAuthor: false
    phabricator:
      url: https://secure.phabricator.com
      token: api-xxxxxxxx
    channels:
      objectTypes: 
        - 'TASK'
      projects:
        default: ""
        1: '#devops'
        2: '#developers'
```