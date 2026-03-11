package main

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/rackerlabs/goclouddns/domains"
	"github.com/rackerlabs/goclouddns/records"
)

func TestNormalizeLegacyArgsRecordCommand(t *testing.T) {
	args := []string{
		"record",
		"2ea5c211-900c-4c91-adb6-78de5464d60a",
		"update",
		"a978cfa7-2b36-46d5-b0c0-17532705f1d0",
		"-data",
		"10.5.19.11",
	}

	got := normalizeLegacyArgs(args)
	want := []string{
		"record",
		"update",
		"2ea5c211-900c-4c91-adb6-78de5464d60a",
		"a978cfa7-2b36-46d5-b0c0-17532705f1d0",
		"-data",
		"10.5.19.11",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeLegacyArgs() = %v, want %v", got, want)
	}
}

func TestNormalizeLegacyArgsLeavesNewSyntaxUntouched(t *testing.T) {
	args := []string{
		"record",
		"update",
		"2ea5c211-900c-4c91-adb6-78de5464d60a",
		"a978cfa7-2b36-46d5-b0c0-17532705f1d0",
		"-data",
		"10.5.19.11",
	}

	got := normalizeLegacyArgs(args)
	if !reflect.DeepEqual(got, args) {
		t.Fatalf("normalizeLegacyArgs() = %v, want %v", got, args)
	}
}

func TestRootVersionFlag(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "clouddns version") {
		t.Fatalf("expected version output, got %q", output)
	}
	if !strings.Contains(output, "commit:") {
		t.Fatalf("expected commit output, got %q", output)
	}
	if !strings.Contains(output, "built:") {
		t.Fatalf("expected build output, got %q", output)
	}
}

func TestInvalidFormatFailsBeforeServiceSetup(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"domain", "list", "--format", "yaml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid format error")
	}
	if !strings.Contains(err.Error(), "unsupported --format") {
		t.Fatalf("expected invalid format error, got %q", err)
	}
}

func TestRecordUpdateRequiresAtLeastOneChangeFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"record", "update", "domid", "recid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing change flag error")
	}
	if !strings.Contains(err.Error(), "specify at least one of --data, --ttl, --priority, --comment") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestRecordUpdateMissingArgsHasUsefulError(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"record", "update"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected argument error")
	}
	if !strings.Contains(err.Error(), "missing required arguments: DOMID and ID") {
		t.Fatalf("unexpected error: %q", err)
	}
	if !strings.Contains(err.Error(), "Usage:\n  clouddns record update DOMID ID") {
		t.Fatalf("unexpected usage error: %q", err)
	}
	if !strings.Contains(err.Error(), "Flags:") {
		t.Fatalf("expected flags in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "--data string") {
		t.Fatalf("expected local flags in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "Global Flags:") {
		t.Fatalf("expected global flags in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "Examples:") {
		t.Fatalf("expected examples in error, got %q", err)
	}
}

func TestRecordListMissingArgsHasFriendlyError(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"record", "list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected argument error")
	}
	if !strings.Contains(err.Error(), "missing required argument: DOMID") {
		t.Fatalf("unexpected error: %q", err)
	}
	if !strings.Contains(err.Error(), "Usage:\n  clouddns record list DOMID") {
		t.Fatalf("unexpected usage error: %q", err)
	}
	if !strings.Contains(err.Error(), "--type string") {
		t.Fatalf("expected --type flag in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "--format string") {
		t.Fatalf("expected --format flag in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "Examples:") {
		t.Fatalf("expected examples in error, got %q", err)
	}
}

func TestRecordUpdateHelpIncludesExamples(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"record", "update", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Examples:") {
		t.Fatalf("expected examples section, got %q", output)
	}
	if !strings.Contains(output, "clouddns record update <domain-id> <record-id>") {
		t.Fatalf("expected update example, got %q", output)
	}
	if !strings.Contains(output, "clouddns record <domain-id> update <record-id>") {
		t.Fatalf("expected legacy syntax example, got %q", output)
	}
}

func TestPrintDomainListsCompactTable(t *testing.T) {
	output := captureStdout(t, func() {
		err := printDomainLists("table", false, []domains.DomainList{
			{
				ID:      "2ea5c211-900c-4c91-adb6-78de5464d60a",
				Name:    "prod.undercloud.rackspace.net",
				Email:   "hostmaster-long-email-address@rackspace.com",
				Updated: "2026-03-11T20:01:52.950Z",
				Created: "2025-03-11T11:41:11.175Z",
			},
		})
		if err != nil {
			t.Fatalf("printDomainLists() returned error: %v", err)
		}
	})

	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") || !strings.Contains(output, "UPDATED") {
		t.Fatalf("expected compact table headers, got %q", output)
	}
	firstLine := strings.Split(strings.TrimSpace(output), "\n")[0]
	if !strings.HasPrefix(firstLine, "ID") {
		t.Fatalf("expected ID-first compact header, got %q", firstLine)
	}
	if strings.Contains(output, "CREATED") {
		t.Fatalf("did not expect CREATED in compact output, got %q", output)
	}
	if strings.Contains(output, "EMAIL") {
		t.Fatalf("did not expect EMAIL in compact output, got %q", output)
	}
	if !strings.Contains(output, "2ea5c211-900c-4c91-adb6-78de5464d60a") {
		t.Fatalf("expected full id, got %q", output)
	}
	if !strings.Contains(output, "2026-03-11 20:01") {
		t.Fatalf("expected compact timestamp, got %q", output)
	}
}

func TestPrintDomainListsWideTable(t *testing.T) {
	output := captureStdout(t, func() {
		err := printDomainLists("table", true, []domains.DomainList{
			{
				ID:      "2ea5c211-900c-4c91-adb6-78de5464d60a",
				Name:    "prod.undercloud.rackspace.net",
				Email:   "hostmaster@rackspace.com",
				Updated: "2026-03-11T20:01:52.950Z",
				Created: "2025-03-11T11:41:11.175Z",
			},
		})
		if err != nil {
			t.Fatalf("printDomainLists() returned error: %v", err)
		}
	})

	if !strings.Contains(output, "CREATED") {
		t.Fatalf("expected wide table header, got %q", output)
	}
	if !strings.Contains(output, "2ea5c211-900c-4c91-adb6-78de5464d60a") {
		t.Fatalf("expected full id in wide output, got %q", output)
	}
}

func TestPrintRecordListsCompactTable(t *testing.T) {
	output := captureStdout(t, func() {
		err := printRecordLists("table", false, []records.RecordList{
			{
				ID:   "a978cfa7-2b36-46d5-b0c0-17532705f1d0",
				Name: "app.prod.undercloud.rackspace.net",
				Type: "A",
				Data: "10.5.19.11",
				TTL:  300,
			},
		})
		if err != nil {
			t.Fatalf("printRecordLists() returned error: %v", err)
		}
	})

	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") || !strings.Contains(output, "TYPE") {
		t.Fatalf("expected compact record headers, got %q", output)
	}
	firstLine := strings.Split(strings.TrimSpace(output), "\n")[0]
	if !strings.HasPrefix(firstLine, "ID") {
		t.Fatalf("expected ID-first compact header, got %q", firstLine)
	}
	if strings.Contains(output, "PRIORITY") {
		t.Fatalf("did not expect PRIORITY in compact output, got %q", output)
	}
	if !strings.Contains(output, "a978cfa7-2b36-46d5-b0c0-17532705f1d0") {
		t.Fatalf("expected full record id, got %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() failed: %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	return string(out)
}
