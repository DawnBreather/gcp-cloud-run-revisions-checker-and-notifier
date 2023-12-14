package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "revisions-checker/common"
	. "revisions-checker/config"
)

// SlackRequestBody is the request payload that Slack expects for an Incoming Webhook.
type SlackRequestBody struct {
	Text string `json:"text"`
}

// SendSlackNotification sends a message to a Slack channel.
func SendSlackNotification(revision Revision, service Service, config Configuration) error {

	msg := fmt.Sprintf(
		":rocket: *New Active Revision Detected!*\n\n"+
			":mag: *Revision Details:*\n"+
			"- *Name:* `%s`\n"+
			"- *Image:* `%s`\n"+
			"- *Creation Time:* `%s`\n"+
			"- *Service:* `%s`\n\n"+
			":pushpin: *Additional Info:*\n"+
			"- *Project ID:* `%s`\n"+
			"- *Region:* `%s`\n\n"+
			":link: *View in Cloud Console:* <%s|Link to Cloud Run Service>\n\n"+
			":loudspeaker: *What's Next?*\n"+
			"- Review the new revision.\n"+
			"- Monitor performance and error rates.\n"+
			"- Ensure that the revision is operating as expected.\n",
		revision.Name, revision.Image, revision.CreationTime.Format(time.RFC1123),
		service.Name, config.ProjectID, config.Region, service.URL)

	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, config.SlackWebhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return fmt.Errorf("Non-ok response returned from Slack")
	}
	defer resp.Body.Close()

	return nil
}
