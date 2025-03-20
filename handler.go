package main

import (
	"bytes"
	"encoding/base64"
	"log/slog"
	"net"
	"net/mail"
	"regexp"

	"google.golang.org/api/gmail/v1"
)

func MailHandler(srv *gmail.Service, senders []*regexp.Regexp, recipients []*regexp.Regexp) func(origin net.Addr, from string, to []string, data []byte) error {
	slog.Info("Handler is ready.")
	return func(origin net.Addr, from string, to []string, data []byte) error {
		msg, err := mail.ReadMessage(bytes.NewReader(data))
		if err != nil {
			slog.Warn("Unable to decode message, discarding!", "error", err)
			return nil
		}
		subject := msg.Header.Get("Subject")
		slog.Info("Received mail", "from", from, "to", to, "subject", subject)
		if !MatchAnyPattern(from, senders) {
			slog.Warn("Ignoring email due to sender restrictions", "from", from)
			return nil
		}
		for _, target := range to {
			if !MatchAnyPattern(target, recipients) {
				slog.Warn("Ignoring email due to recipient restrictions", "to", target)
				return nil
			}
		}
		_, err = srv.Users.Messages.Send("me", &gmail.Message{
			Raw: base64.StdEncoding.EncodeToString(data),
		}).Do()
		if err != nil {
			slog.Warn("Unable to send email", "error", err)
			return err
		} else {
			slog.Info("Message sent successfully.", "to", to)
		}
		return nil
	}
}
