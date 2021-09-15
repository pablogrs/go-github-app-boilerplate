package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.hpe.com/sharkysharks/go-github-app-boilerplate/config"
)

type WebhookAPIRequest struct {
	Action       string              `json:"action"`
	Installation AppInstallation     `json:"installation"`
	Issue        github.Issue        `json:"issue"`
	IssueComment github.IssueComment `json:"comment"`
	Repo         github.Repository   `json:"repository"`
}

type AppInstallation struct {
	Id int64 `json:"id"`
}

var ghClient *github.Client
var conf *config.Config
var Payload *WebhookAPIRequest

func SetConfig(config *config.Config) {
	conf = config
}

// validate payload webhook signature
func ValidatePayload(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := github.ValidatePayload(r, []byte(conf.GithubApp.GithubWebhookSecret))
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		pd, err := getWebhookAPIRequest(p)
		if err != nil {
			log.Error("Error unmarshalling json payload: ", err)
			return
		}
		Payload = pd
		next.ServeHTTP(w, r)
	})
}

// helper function to convert json payload to struct
func getWebhookAPIRequest(body []byte) (*WebhookAPIRequest, error) {
	var wh = new(WebhookAPIRequest)
	err := json.Unmarshal(body, &wh)
	if err != nil {
		return nil, err
	}
	return wh, nil
}

// authenticate as Github App
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := initClient(Payload.Installation.Id, conf)
		if err != nil {
			log.Error("Error initializing client: ", err)
			return
		}
		ghClient = c

		next.ServeHTTP(w, r)
	})
}

func initClient(installationId int64, conf *config.Config) (*github.Client, error) {
	tr := http.DefaultTransport
	itr, err := ghinstallation.New(
		tr,
		conf.GithubApp.GithubAppIdentifier,
		installationId,
		[]byte(conf.GithubApp.GithubPrivateKey),
	)
	if err != nil {
		return nil, err
	}

	c := github.NewClient(&http.Client{Transport: itr})

	return c, nil
}
