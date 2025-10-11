package imapclient_test

import (
	"os"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

func TestClient_Notify(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("NOTIFY test requires GOIMAP_TEST_DOVECOT=1")
	}

	existsCh := make(chan uint32, 1)

	// Clients needs a UnilateralDataHandler configured, so set it up manually.
	conn, server := newDovecotClientServerPair(t)
	defer server.Close()

	options := &imapclient.Options{
		UnilateralDataHandler: &imapclient.UnilateralDataHandler{
			Expunge: func(seqNum uint32) {
				// Not testing expunge in this test
			},
			Mailbox: func(data *imapclient.UnilateralDataMailbox) {
				if data.NumMessages != nil {
					select {
					case existsCh <- *data.NumMessages:
					default:
					}
				}
			},
		},
	}
	if testing.Verbose() {
		options.DebugWriter = os.Stderr
	}

	client := imapclient.New(conn, options)
	defer client.Close()

	// (Dovecot connections are pre-authenticated)
	if err := client.WaitGreeting(); err != nil {
		t.Fatalf("WaitGreeting() = %v", err)
	}

	// Append initial message to INBOX.
	appendCmd := client.Append("INBOX", int64(len(simpleRawMessage)), nil)
	appendCmd.Write([]byte(simpleRawMessage))
	appendCmd.Close()
	if _, err := appendCmd.Wait(); err != nil {
		t.Fatalf("Initial Append() = %v", err)
	}

	selectData, err := client.Select("INBOX", nil).Wait()
	if err != nil {
		t.Fatalf("Select() = %v", err)
	}
	initialExists := selectData.NumMessages

	notifyOptions := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSelected,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMessageExpunge,
				},
			},
		},
	}
	cmd, err := client.Notify(notifyOptions)
	if err != nil {
		t.Fatalf("Notify() = %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected non-nil NotifyCommand")
	}

	// Append a new message to INBOX (we should get a NOTIFY event for it).
	testMessage := `From: sender@example.com
To: recipient@example.com
Subject: Test NOTIFY

This is a test message for NOTIFY.
`
	appendCmd = client.Append("INBOX", int64(len(testMessage)), nil)
	appendCmd.Write([]byte(testMessage))
	appendCmd.Close()
	if _, err := appendCmd.Wait(); err != nil {
		t.Fatalf("Append() = %v", err)
	}

	// Wait for the EXISTS notification (with timeout)
	select {
	case count := <-existsCh:
		if count <= initialExists {
			t.Errorf("Expected EXISTS count > %d, got %d", initialExists, count)
		}
		t.Logf("Received EXISTS notification: %d messages (was %d)", count, initialExists)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for EXISTS notification")
	}
}

func TestClient_NotifyNone(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("Skipping NOTIFY test - requires GOIMAP_TEST_DOVECOT=1")
	}

	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	_, err := client.Notify(nil)
	if err != nil {
		t.Fatalf("NotifyNone() = %v", err)
	}
}

func TestClient_NotifyMultiple(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("Skipping NOTIFY test - requires GOIMAP_TEST_DOVECOT=1")
	}

	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	// Test NOTIFY with multiple items
	// Note: Dovecot doesn't support STATUS with message events
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSelected,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMessageExpunge,
				},
			},
			{
				MailboxSpec: imap.NotifyMailboxSpecPersonal,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMailboxName,
					imap.NotifyEventSubscriptionChange,
				},
			},
		},
	}

	cmd, err := client.Notify(options)
	if err != nil {
		t.Fatalf("Notify() = %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil NotifyCommand")
	}
}

// TestClient_NotifyPersonalMailboxes tests NOTIFY for personal mailboxes
func TestClient_NotifyPersonalMailboxes(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("Skipping NOTIFY test - requires GOIMAP_TEST_DOVECOT=1")
	}

	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	// Note: Dovecot doesn't support message events with PERSONAL spec
	// Only mailbox events (MailboxName, SubscriptionChange) seem to work.
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecPersonal,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMailboxName,
					imap.NotifyEventSubscriptionChange,
				},
			},
		},
	}

	cmd, err := client.Notify(options)
	if err != nil {
		t.Fatalf("Notify() = %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil NotifyCommand")
	}
}

// TestClient_NotifySubtree tests NOTIFY for mailbox subtrees
func TestClient_NotifySubtree(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("Skipping NOTIFY test - requires GOIMAP_TEST_DOVECOT=1")
	}

	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	// Request notifications for INBOX subtree
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				Mailboxes: []string{"INBOX"},
				Subtree:   true,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMailboxName,
					imap.NotifyEventSubscriptionChange,
				},
			},
		},
	}

	cmd, err := client.Notify(options)
	if err != nil {
		t.Fatalf("Notify() = %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil NotifyCommand")
	}
}

// TestClient_NotifyMailboxes tests NOTIFY for specific mailboxes with message events
func TestClient_NotifyMailboxes(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("Skipping NOTIFY test - requires GOIMAP_TEST_DOVECOT=1")
	}

	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	// Request notifications for specific mailboxes with SUBTREE
	// Note: Dovecot requires SUBTREE for explicit mailbox specifications
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				Mailboxes: []string{"INBOX"},
				Subtree:   true,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMailboxName,
					imap.NotifyEventSubscriptionChange,
				},
			},
		},
	}

	cmd, err := client.Notify(options)
	if err != nil {
		t.Fatalf("Notify() = %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil NotifyCommand")
	}
}

// TestClient_NotifySelectedDelayed tests NOTIFY with SELECTED-DELAYED for safe MSN usage
func TestClient_NotifySelectedDelayed(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("Skipping NOTIFY test - requires GOIMAP_TEST_DOVECOT=1")
	}

	client, server := newClientServerPair(t, imap.ConnStateSelected)
	defer client.Close()
	defer server.Close()

	// Request notifications with SELECTED-DELAYED to defer expunge notifications
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSelectedDelayed,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMessageExpunge,
				},
			},
		},
	}

	cmd, err := client.Notify(options)
	if err != nil {
		t.Fatalf("Notify() = %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil NotifyCommand")
	}
}

// TestClient_NotifySequence tests a sequence of NOTIFY commands
func TestClient_NotifySequence(t *testing.T) {
	if os.Getenv("GOIMAP_TEST_DOVECOT") != "1" {
		t.Skip("Skipping NOTIFY test - requires GOIMAP_TEST_DOVECOT=1")
	}

	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	// First NOTIFY command
	options1 := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSelected,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMessageExpunge,
				},
			},
		},
	}

	cmd1, err := client.Notify(options1)
	if err != nil {
		t.Fatalf("First Notify() = %v", err)
	}
	if cmd1 == nil {
		t.Fatal("Expected non-nil NotifyCommand from first call")
	}

	// Replace with different NOTIFY settings
	options2 := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecPersonal,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMailboxName,
					imap.NotifyEventSubscriptionChange,
				},
			},
		},
	}

	cmd2, err := client.Notify(options2)
	if err != nil {
		t.Fatalf("Second Notify() = %v", err)
	}
	if cmd2 == nil {
		t.Fatal("Expected non-nil NotifyCommand from second call")
	}

	// Disable all notifications
	_, err = client.Notify(nil)
	if err != nil {
		t.Fatalf("NotifyNone() = %v", err)
	}
}
