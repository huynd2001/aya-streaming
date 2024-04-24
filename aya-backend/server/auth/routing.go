package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"golang.org/x/oauth2"
)

type VerificationEvent struct {
	Code  string `schema:"code"`
	State string `schema:"state"`
	Scope string `schema:"scope"`
}

type Workflow struct {
	tokenSourceCh  chan oauth2.TokenSource
	verificationCh chan VerificationEvent
	authURL        *string
	verified       bool
}

func NewWorkflow() *Workflow {
	tokenSourceCh := make(chan oauth2.TokenSource)
	verificationCh := make(chan VerificationEvent)
	return &Workflow{
		tokenSourceCh:  tokenSourceCh,
		verificationCh: verificationCh,
		verified:       false,
		authURL:        nil,
	}
}

func (workflow *Workflow) TokenSourceCh() chan oauth2.TokenSource {
	return workflow.tokenSourceCh
}

func (workflow *Workflow) SetupAuth(
	oauthConfig oauth2.Config,
	promptURL string,
) {

	verifier := oauth2.GenerateVerifier()

	workflow.verified = false
	workflow.authURL = nil
	workflow.tokenSourceCh = make(chan oauth2.TokenSource)
	workflow.verificationCh = make(chan VerificationEvent)

	var stateStr string
	stateUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Println("Error generating uuid. Opt for default state string")
		stateStr = "state"
	} else {
		stateStr = stateUUID.String()
	}

	url := oauthConfig.AuthCodeURL(stateStr, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	workflow.authURL = &url
	fmt.Printf("Visit the link to start the auth process:\n%s\n", promptURL)
	fmt.Printf("If the thing is not working, try the following link instead:\n%s\n", url)

	// Wait for the code challenge
	go func() {
		for {
			verificationRes := <-workflow.verificationCh
			if verificationRes.State != stateStr {
				fmt.Println("Invalid State received!")
				continue
			}

			ctx := context.Background()

			oauth2Token, err := oauthConfig.Exchange(ctx, verificationRes.Code, oauth2.VerifierOption(verifier))
			if err != nil {
				// Wrong code. Continue waiting
				fmt.Printf("Error during code exchange: %s\n", err.Error())
				continue
			}

			workflow.verified = true
			workflow.authURL = nil
			workflow.tokenSourceCh <- oauthConfig.TokenSource(context.Background(), oauth2Token)
			close(workflow.tokenSourceCh)
			break
		}
	}()

}

func (workflow *Workflow) SetUpRedirectAndCodeChallenge(
	redirectRoute *mux.Router,
	callbackRoute *mux.Router,
) {
	redirectRoute.PathPrefix("").HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if workflow.authURL != nil {
			http.Redirect(writer, req, *workflow.authURL, http.StatusFound)
			return
		} else if workflow.verified {
			http.Error(writer, "Already verified", http.StatusUnauthorized)
			return
		} else {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			return
		}
	})

	callbackRoute.PathPrefix("").HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if workflow.verified {
			http.Error(writer, "Not authorized", http.StatusUnauthorized)
			return
		} else {

			reqQuery := req.URL.Query()
			var decoder = schema.NewDecoder()
			var verifyEvent VerificationEvent

			err := decoder.Decode(&verifyEvent, reqQuery)
			if err != nil {
				fmt.Printf("Error during decoding:%s\n", err)
				http.Error(writer, "Bad request", http.StatusBadRequest)
				return
			}

			workflow.verificationCh <- verifyEvent
			close(workflow.verificationCh)
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte("Code and State has been sent to server. Check the log for more information"))
			return
		}
	})
}
