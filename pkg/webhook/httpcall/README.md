# Send http request

Can be used to send http requests to a webhook. For example slack notifications.

```yaml
webhooks: 
- provider: httpcall
  ids:
  - cluster:namespace
  config:
    url: https://hooks.slack.com/services/xxx/xxx/xxx
    headers:
      Content-Type: application/json
    body: |
      {
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": ":ghost: {{ .Message.Namespace }} <https://kubernetes-manager.com/{{ .Message.Cluster }}:{{ .Message.Namespace }}/settings|{{ .Message.Reason }}>"
            }
          }
        ]
      }
```
