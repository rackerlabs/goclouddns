package goclouddns

import (
	"encoding/json"
	"testing"

	"github.com/gophercloud/gophercloud/v2"
)

func TestAsyncMessage_BasicFields(t *testing.T) {
	message := AsyncMessage{
		CallbackURL: "https://example.com/callback",
		JobID:       "job-123",
		Status:      "RUNNING",
	}

	if message.CallbackURL != "https://example.com/callback" {
		t.Errorf("expected CallbackURL to be set correctly")
	}
	if message.JobID != "job-123" {
		t.Errorf("expected JobID to be set correctly")
	}
	if message.Status != "RUNNING" {
		t.Errorf("expected Status to be set correctly")
	}
}

func TestAsyncMessage_JSONMarshaling(t *testing.T) {
	message := AsyncMessage{
		CallbackURL: "https://example.com/callback",
		JobID:       "job-123",
		Request:     "POST",
		Status:      "COMPLETED",
		Response:    map[string]interface{}{"id": "123"},
		Error:       map[string]interface{}{"code": 400},
		RequestURL:  "https://example.com/domains",
		Verb:        "POST",
	}

	// Test marshaling
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("failed to marshal AsyncMessage: %v", err)
	}

	// Test unmarshaling
	var unmarshaled AsyncMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal AsyncMessage: %v", err)
	}

	if unmarshaled.CallbackURL != message.CallbackURL {
		t.Errorf("CallbackURL mismatch after JSON round-trip")
	}
	if unmarshaled.JobID != message.JobID {
		t.Errorf("JobID mismatch after JSON round-trip")
	}
	if unmarshaled.Status != message.Status {
		t.Errorf("Status mismatch after JSON round-trip")
	}
}

func TestNewCloudDNS_BasicFunctionality(t *testing.T) {
	client := &gophercloud.ProviderClient{
		EndpointLocator: func(opts gophercloud.EndpointOpts) (string, error) {
			return "https://dns.api.rackspacecloud.com/v1.0/123456", nil
		},
	}

	sc, err := NewCloudDNS(client, gophercloud.EndpointOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sc == nil {
		t.Fatalf("expected service client but got nil")
	}

	if sc.Type != "rax:dns" {
		t.Errorf("expected service type 'rax:dns', got %q", sc.Type)
	}

	if sc.ProviderClient != client {
		t.Errorf("expected provider client to be set correctly")
	}

	if sc.Endpoint == "" {
		t.Errorf("expected endpoint to be set")
	}
}

func TestNewCloudDNS_EndpointError(t *testing.T) {
	client := &gophercloud.ProviderClient{
		EndpointLocator: func(opts gophercloud.EndpointOpts) (string, error) {
			return "", gophercloud.ErrEndpointNotFound{}
		},
	}

	_, err := NewCloudDNS(client, gophercloud.EndpointOpts{})
	if err == nil {
		t.Errorf("expected endpoint error but got none")
	}
}
