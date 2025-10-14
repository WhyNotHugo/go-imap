package imapclient_test

import (
	"testing"

	"github.com/emersion/go-imap/v2"
)

func TestClient_EnableQResync(t *testing.T) {
	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	// Enable QRESYNC
	data, err := client.Enable(imap.CapQResync).Wait()
	if err != nil {
		t.Fatalf("Enable(QRESYNC) failed: %v", err)
	}

	if !data.Caps.Has(imap.CapQResync) {
		t.Errorf("QRESYNC was not enabled")
	}

	// QRESYNC should also enable CONDSTORE
	if !data.Caps.Has(imap.CapCondStore) {
		t.Errorf("CONDSTORE was not enabled (should be implied by QRESYNC)")
	}
}

func TestClient_SelectQResync(t *testing.T) {
	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	// Enable QRESYNC first
	_, err := client.Enable(imap.CapQResync).Wait()
	if err != nil {
		t.Fatalf("Enable(QRESYNC) failed: %v", err)
	}

	// Select with QRESYNC parameters
	knownUIDs := imap.UIDSetNum(1, 5)
	selectData, err := client.Select("INBOX", &imap.SelectOptions{
		QResync: &imap.SelectQResyncOptions{
			UIDValidity: 123456,
			ModSeq:      789,
			KnownUIDs:   &knownUIDs,
		},
	}).Wait()
	if err != nil {
		t.Fatalf("Select with QRESYNC failed: %v", err)
	}

	if selectData == nil {
		t.Fatal("Select returned nil data")
	}
}

func TestClient_FetchVanished(t *testing.T) {
	client, server := newClientServerPair(t, imap.ConnStateSelected)
	defer client.Close()
	defer server.Close()

	// Enable QRESYNC first
	_, err := client.Enable(imap.CapQResync).Wait()
	if err != nil {
		t.Fatalf("Enable(QRESYNC) failed: %v", err)
	}

	// Fetch with VANISHED modifier
	uidSet := imap.UIDSetNum(1, 100)
	fetchCmd := client.Fetch(uidSet, &imap.FetchOptions{
		UID:          true,
		ChangedSince: 100,
		Vanished:     true,
	})

	// Just ensure the command completes without error
	// (we can't easily test the actual VANISHED response without a full server)
	if err := fetchCmd.Close(); err != nil {
		t.Fatalf("Fetch with VANISHED failed: %v", err)
	}
}
