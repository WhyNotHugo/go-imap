package imapclient

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/internal/imapwire"
)

func encodeToString(options *imap.NotifyOptions) (string, error) {
	buf := &bytes.Buffer{}
	bw := bufio.NewWriter(buf)
	enc := imapwire.NewEncoder(bw, imapwire.ConnSideClient)

	if err := encodeNotifyOptions(enc, options); err != nil {
		return "", err
	}

	enc.CRLF()
	bw.Flush()

	return buf.String(), nil
}

func TestEncodeNotifyOptions_None(t *testing.T) {
	result, err := encodeToString(nil)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " NONE\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_EmptyItems(t *testing.T) {
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " NONE\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_Selected(t *testing.T) {
	options := &imap.NotifyOptions{
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
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (SELECTED (MessageNew MessageExpunge))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_SelectedDelayed(t *testing.T) {
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
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (SELECTED-DELAYED (MessageNew MessageExpunge))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_Personal(t *testing.T) {
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
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (PERSONAL (MailboxName SubscriptionChange))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_Inboxes(t *testing.T) {
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecInboxes,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (INBOXES (MessageNew))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_Subscribed(t *testing.T) {
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSubscribed,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMailboxName,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (SUBSCRIBED (MessageNew MailboxName))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_Subtree(t *testing.T) {
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				Subtree:   true,
				Mailboxes: []string{"INBOX", "Lists"},
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (SUBTREE (INBOX \"Lists\") (MessageNew))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_MailboxList(t *testing.T) {
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				Mailboxes: []string{"INBOX", "Sent"},
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMessageExpunge,
					imap.NotifyEventFlagChange,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET ((INBOX \"Sent\") (MessageNew MessageExpunge FlagChange))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_StatusIndicator(t *testing.T) {
	options := &imap.NotifyOptions{
		Status: true,
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
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (STATUS) (SELECTED (MessageNew MessageExpunge))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_MultipleItems(t *testing.T) {
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
			{
				MailboxSpec: imap.NotifyMailboxSpecInboxes,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (SELECTED (MessageNew MessageExpunge)) (PERSONAL (MailboxName SubscriptionChange)) (INBOXES (MessageNew))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_AllEvents(t *testing.T) {
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSelected,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMessageExpunge,
					imap.NotifyEventFlagChange,
					imap.NotifyEventAnnotationChange,
					imap.NotifyEventMailboxName,
					imap.NotifyEventSubscriptionChange,
					imap.NotifyEventMailboxMetadataChange,
					imap.NotifyEventServerMetadataChange,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (SELECTED (MessageNew MessageExpunge FlagChange AnnotationChange MailboxName SubscriptionChange MailboxMetadataChange ServerMetadataChange))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_NoEvents(t *testing.T) {
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSelected,
				Events:      []imap.NotifyEvent{},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (SELECTED)\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_InvalidItem(t *testing.T) {
	// Items with neither MailboxSpec nor Mailboxes should return an error
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				// Invalid: no mailbox spec or mailboxes
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
				},
			},
		},
	}
	_, err := encodeToString(options)
	if err == nil {
		t.Fatal("Expected error for invalid NOTIFY item, got nil")
	}

	expectedMsg := "invalid NOTIFY item: must specify either MailboxSpec or Mailboxes"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestEncodeNotifyOptions_ComplexMixed(t *testing.T) {
	options := &imap.NotifyOptions{
		Status: true,
		Items: []imap.NotifyItem{
			{
				MailboxSpec: imap.NotifyMailboxSpecSelected,
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
					imap.NotifyEventMessageExpunge,
				},
			},
			{
				Subtree:   true,
				Mailboxes: []string{"INBOX"},
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
				},
			},
			{
				Mailboxes: []string{"Drafts", "Sent"},
				Events: []imap.NotifyEvent{
					imap.NotifyEventFlagChange,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET (STATUS) (SELECTED (MessageNew MessageExpunge)) (SUBTREE (INBOX) (MessageNew)) ((\"Drafts\" \"Sent\") (FlagChange))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeNotifyOptions_MailboxWithSpecialChars(t *testing.T) {
	// Test mailbox names that require quoting
	options := &imap.NotifyOptions{
		Items: []imap.NotifyItem{
			{
				Mailboxes: []string{"INBOX", "Foo Bar", "Test&Mailbox"},
				Events: []imap.NotifyEvent{
					imap.NotifyEventMessageNew,
				},
			},
		},
	}
	result, err := encodeToString(options)
	if err != nil {
		t.Fatalf("encodeToString() error = %v", err)
	}
	expected := " SET ((INBOX \"Foo Bar\" \"Test&-Mailbox\") (MessageNew))\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
