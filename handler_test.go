package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func fakeMailHandler(t *testing.T, m *MockGmailService, senders []*regexp.Regexp, recipients []*regexp.Regexp) func(origin net.Addr, from string, to []string, data []byte) error {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if m.ShouldFail {
			http.Error(w, "Failure!!!", http.StatusBadRequest)
			return
		}

		// Parse request body
		var msg gmail.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		m.SentMessages = append(m.SentMessages, &msg)

		data, err := base64.StdEncoding.DecodeString(msg.Raw)
		if err != nil {
			http.Error(w, "Unable to decode message", http.StatusInternalServerError)
			return
		}
		fmt.Println(string(data))

		// Validate required fields
		if false /*emailReq.To == "" || emailReq.Subject == "" || emailReq.Body == ""*/ {
			http.Error(w, "Missing required fields: to, subject, body", http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Email sent successfully",
		})
	}))
	srv, err := gmail.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))
	if err != nil {
		t.Fatalf("unable to create client: %v", err)
	}
	return MailHandler(srv, senders, recipients)
}

// MockAddr implements net.Addr interface for testing
type MockAddr struct{}

func (m MockAddr) Network() string { return "tcp" }
func (m MockAddr) String() string  { return "192.0.2.1:25" }

// MockGmailService is assumed to be your existing fake
type MockGmailService struct {
	SentMessages []*gmail.Message
	ShouldFail   bool
}

func TestMailHandler(t *testing.T) {
	tests := []struct {
		name              string
		senders           string
		recipients        string
		from              string
		to                []string
		serviceShouldFail bool
		expectSend        bool
	}{
		{
			name:       "Valid sender and recipient",
			senders:    `.*@example\.com`,
			recipients: `.*@destination\.com`,
			from:       "user@example.com",
			to:         []string{"recipient@destination.com"},
			expectSend: true,
		},
		{
			name:       "Invalid sender",
			senders:    `.*@example\.com`,
			recipients: `.*@destination\.com`,
			from:       "user@unknown.com",
			to:         []string{"recipient@destination.com"},
			expectSend: false,
		},
		{
			name:       "Invalid recipient",
			senders:    `.*@example\.com`,
			recipients: `.*@destination\.com`,
			from:       "user@example.com",
			to:         []string{"recipient@unknown.com"},
			expectSend: false,
		},
		{
			name:       "Multiple recipients with one invalid",
			senders:    `.*@example\.com`,
			recipients: `.*@destination\.com`,
			from:       "user@example.com",
			to:         []string{"recipient@destination.com", "other@unknown.com"},
			expectSend: false,
		},
		{
			name:              "Service failure",
			senders:           `.*@example\.com`,
			recipients:        `.*@destination\.com`,
			from:              "user@example.com",
			to:                []string{"recipient@destination.com"},
			serviceShouldFail: true,
			expectSend:        true, // The handler will attempt to send
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Compile regexps
			senderRegexps, _ := ConvertToRegexPatterns(tc.senders)
			recipientRegexps, _ := ConvertToRegexPatterns(tc.recipients)

			// Create mock service
			mockService := &MockGmailService{
				SentMessages: []*gmail.Message{},
				ShouldFail:   tc.serviceShouldFail,
			}

			// Create handler
			handler := fakeMailHandler(t, mockService, senderRegexps, recipientRegexps)

			// Create test email
			emailData := createTestEmail(tc.from, tc.to)

			// Call handler
			err := handler(MockAddr{}, tc.from, tc.to, emailData)

			// Check results
			if tc.expectSend && !tc.serviceShouldFail {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if len(mockService.SentMessages) != 1 {
					t.Errorf("Expected 1 message to be sent, got: %d", len(mockService.SentMessages))
				}
			} else if tc.expectSend && tc.serviceShouldFail {
				if err == nil {
					t.Error("Expected error when service fails, got nil")
				}
				if len(mockService.SentMessages) != 0 {
					t.Errorf("Expected 0 messages to be sent when service fails, got: %d", len(mockService.SentMessages))
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for rejected email, got: %v", err)
				}
				if len(mockService.SentMessages) != 0 {
					t.Errorf("Expected 0 messages to be sent for rejected email, got: %d", len(mockService.SentMessages))
				}
			}
		})
	}
}

// Helper function to create a simple test email
func createTestEmail(from string, to []string) []byte {
	var buf bytes.Buffer

	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + to[0] + "\r\n") // Simplified for testing
	buf.WriteString("Subject: Test Email\r\n")
	buf.WriteString("Content-Type: text/plain\r\n")
	buf.WriteString("\r\n")
	buf.WriteString("This is a test email body.")

	return buf.Bytes()
}
