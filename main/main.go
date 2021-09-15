package main

import (
	"net/http"
	"strings"

	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"github.hpe.com/sharkysharks/go-github-app-boilerplate/config"
	"github.hpe.com/sharkysharks/go-github-app-boilerplate/middleware"
)

/*
	Add top level GitHub request payload keys to this struct based on the events the app is subscribing to and pull in
	the type from the go-github/github library
*/

func init() {
	log.SetFormatter(&log.TextFormatter{})

	c, err := config.ReadConfig()
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}

	middleware.SetConfig(c)
}

func main() {
	// health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// middleware
	stdChain := alice.New(middleware.ValidatePayload, middleware.Authenticate)

	// main application endpoint
	http.Handle("/", stdChain.Then(http.HandlerFunc(app)))

	log.Info("Server listening...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

// application
func app(w http.ResponseWriter, r *http.Request) {
	/*
		match the requestID to the webhook event logs under the github app settings in the github web console
		this is helpful for debugging webhook events
	*/
	requestID := r.Header.Get("X-GitHub-Delivery")
	/*
		this section is where webhook events will be received after passing through middleware validation
		below is an example of handling a comment on a pull request
		to work, the GitHub app would need to be configured in GitHub to subscribe to Issue Comment creation events
	*/
	if r.Method == "POST" {
		switch githubEvent := r.Header.Get("X-GitHub-Event"); githubEvent {
		case "issue_comment":
			if middleware.Payload.Action == "created" {
				comment := strings.TrimSpace(*middleware.Payload.IssueComment.Body)
				switch comment {
				case "run all tests":
					log.Info("Received event to run all tests")
					//	execute some code here based on receiving a comment on a pull request
				}
			}
		case "installation":
			if middleware.Payload.Action == "created" {
				log.Info("Received installation request")
			}
		default:
			log.Error("No handler for event type: ", githubEvent, "\nRequest ID: ", requestID)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	} else {
		log.Error("Method not allowed: ", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
