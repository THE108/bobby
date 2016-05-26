package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	sc "github.com/nlopes/slack"
)

type Client struct {
	cli *sc.Client
}

func NewClient(token string) *Client {
	return &Client{
		cli: sc.New(token),
	}
}

func (this *Client) SendMessage(channelID, text string) error {
	_, _, err := this.cli.PostMessage(channelID, text, sc.PostMessageParameters{
		AsUser: true,
	})
	return err
}

func (this *Client) SendPostponedMessage(responseURL, message string) error {
	fmt.Printf("responseURL: %s message: %s\n", responseURL, message)

	requestBody, err := json.Marshal(&SlackResult{Text: message})
	if err != nil {
		return err
	}

	resp, err := http.Post(responseURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %s", resp.Status)
	}
	return nil
}

// SlackResult holds the result of processing the command.  json encoding is the `payload`
// message to a slack incoming hook integration.
type SlackResult struct {
	Text        string             `json:"text"`
	Username    string             `json:"username,omitempty"`
	IconUrl     string             `json:"icon_url,omitempty"`
	IconEmoji   string             `json:"icon_emoji,omitempty"`
	Channel     string             `json:"channel,omitempty"`
	Attachments []*SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment is a message attachment
type SlackAttachment struct {
	ImageUrl string `json:"image_url,omitempty"`
	ThumbUrl string `json:"thumb_url,omitempty"`
	Text     string `json:"text,omitempty"`
	Fallback string `json:"fallback,omitempty"`
}
