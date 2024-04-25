package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/mhale/smtpd"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config, tokFile string) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = tokenFromWeb(ctx, config)
		saveToken(tokFile, tok)
	}
	return config.Client(ctx, tok)
}

// Request a token from the web, then returns the retrieved token.
func tokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1><p>Authorized.  You can close this window.</p>")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState)
	log.Printf("Authorize this app at: %s", authURL)
	code := <-ch
	log.Printf("Got code: %s", code)

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}
	return token
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func mailHandler(srv *gmail.Service) func(origin net.Addr, from string, to []string, data []byte) error {
	slog.Info("Handler is ready.")
	return func(origin net.Addr, from string, to []string, data []byte) error {
		msg, err := mail.ReadMessage(bytes.NewReader(data))
		if err != nil {
			slog.Warn("Unable to decode message", "error", err)
			return nil
		}
		subject := msg.Header.Get("Subject")
		slog.Info("Received mail", "from", from, "to", to, "subject", subject)
		_, err = srv.Users.Messages.Send("me", &gmail.Message{
			Raw: base64.StdEncoding.EncodeToString(data),
		}).Do()
		if err != nil {
			slog.Warn("Unable to send email", "error", err)
		} else {
			slog.Info("Message sent successfully.", "to", to)
		}
		return nil
	}
}

type options struct {
	SmtpListen  string `env:"SMTP_LISTEN"`
	Credentials string `env:"CREDENTIALS_JSON"`
	Token       string `env:"TOKEN_JSON"`
}

func main() {
	cfg := options{
		SmtpListen:  ":2525",
		Credentials: "credentials.json",
		Token:       "token.json",
	}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Configuration error", err)
		panic(err)
	}

	ctx := context.Background()
	b, err := os.ReadFile(cfg.Credentials)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(context.Background(), config, cfg.Token)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	if err := smtpd.ListenAndServe(cfg.SmtpListen, mailHandler(srv), "smtp-sidecar", ""); err != nil {
		slog.Error("SMTP Listen error", err)
		panic(err)
	}
}
