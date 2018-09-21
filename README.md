# Phabrick
A Slack Bot for Phabricator

This bot will use the latest API [Herald](https://secure.phabricator.com/herald/) to monitor events 
and send webhooks to the server. The Phabrick server will receive the requests from Herald then
use Conduit API to query responding detail and send notification to Slack

## Installation

* Uses Helm Chart to install on Kubernetes
* Configure Slack Token, Phabricator URL and Token
* Configure mapping between Phabricator project ID and Slack channels
  * You can use `default` to send all unmatched project ID to this Slack channels
  

## phabrick.yaml configuration

Example

```
    slack:
      token: xoxb-xxxxxxx
      username: wall-e
      showAssignee: false
      showAuthor: false
    phabricator:
      url: https://phabricator.omise.co
      token: api-xxxxxxxx
    channels:
      objectTypes: 
        - 'TASK'
      projects:
        default: ""
        11: '#devops-notices'
        25: '#paym-offsite'
        28: '#plugin-woocommerce'
        51: '#devops-notices'
        56: '#func-website'
        64: '#devops-notices'
        67: '#plugin-prestashop'
        68: '#plugin-prestashop'
        77: '#func-reconciliation'
        95: '#func-contracts'
        100: '#intg-omisejs'
        103: '#func-dashboard'
        117: '#func-kyc'
        120: '#func-scheduling'
        121: '#paym-cc-installments'
        128: '#paym-itmx-th'
        143: '#paym-tesco-lotus'
        149: '#defense'
        152: '#func-source'
        154: '#devops-notices'
        155: '#devops-notices'
        156: '#devops-notices'
        163: '#devops-notices'
        165: '#func-reconciliation'
        166: '#paym-alipay'
        167: '#paym-conv-store'
        168: '#plugin-shopify'
        172: '#proj-omise-pay'
        178: '#plugin-magento'
        180: '#plugin-magento'
        186: '#approvals'
```
