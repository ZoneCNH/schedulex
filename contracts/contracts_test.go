package contracts

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ZoneCNH/xlib-standard/pkg/templatex"
)

type schemaProperty struct {
	Type    string   `json:"type"`
	Enum    []string `json:"enum"`
	Minimum *int     `json:"minimum"`
}

type objectSchema struct {
	Required   []string                  `json:"required"`
	Properties map[string]schemaProperty `json:"properties"`
}

func TestErrorKindContractMatchesPublicConstants(t *testing.T) {
	schema := readSchema(t, "error.schema.json")

	expected := sortedStrings(
		string(templatex.ErrorKindConfig),
		string(templatex.ErrorKindValidation),
		string(templatex.ErrorKindConnection),
		string(templatex.ErrorKindUnavailable),
		string(templatex.ErrorKindTimeout),
		string(templatex.ErrorKindAuth),
		string(templatex.ErrorKindConflict),
		string(templatex.ErrorKindRateLimit),
		string(templatex.ErrorKindInternal),
	)
	actual := sortedStrings(schema.Properties["kind"].Enum...)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("error kind contract drift:\nactual:   %#v\nexpected: %#v", actual, expected)
	}
	requireFields(t, schema.Required, "kind", "op", "message", "retryable")
}

func TestHealthStatusContractMatchesPublicConstants(t *testing.T) {
	schema := readSchema(t, "health.schema.json")

	expected := sortedStrings(
		string(templatex.HealthHealthy),
		string(templatex.HealthDegraded),
		string(templatex.HealthUnhealthy),
	)
	actual := sortedStrings(schema.Properties["status"].Enum...)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("health status contract drift:\nactual:   %#v\nexpected: %#v", actual, expected)
	}
	requireFields(t, schema.Required, "name", "status", "checked_at")
}

func TestConfigContractMatchesPublicConfig(t *testing.T) {
	schema := readSchema(t, "config.schema.json")
	requireFields(t, schema.Required, "name")

	configType := reflect.TypeOf(templatex.Config{})
	requireSchemaFieldMapsToStructField(t, schema, configType, "name", "Name", "string")
	requireSchemaFieldMapsToStructField(t, schema, configType, "timeout_ms", "Timeout", "integer")
	requireSchemaFieldMapsToStructField(t, schema, configType, "secret", "Secret", "string")

	if timeoutField, ok := configType.FieldByName("Timeout"); !ok || timeoutField.Type != reflect.TypeOf(time.Duration(0)) {
		t.Fatalf("Config.Timeout must remain time.Duration, got %v", timeoutField.Type)
	}
	if minimum := schema.Properties["timeout_ms"].Minimum; minimum == nil || *minimum != 0 {
		t.Fatalf("timeout_ms must define minimum 0, got %#v", minimum)
	}
}

func TestMetricsContractDocumentsPublicConstants(t *testing.T) {
	content, err := os.ReadFile("metrics.md")
	if err != nil {
		t.Fatalf("read metrics contract: %v", err)
	}
	text := string(content)
	for _, metric := range []string{
		templatex.MetricClientCreatedTotal,
		templatex.MetricClientClosedTotal,
		templatex.MetricClientErrorsTotal,
		templatex.MetricClientHealthStatus,
		templatex.MetricClientHealthLatencyMS,
		templatex.MetricClientRequestsTotal,
		templatex.MetricClientRequestDurationSeconds,
		templatex.MetricClientRetriesTotal,
		templatex.MetricClientInflight,
	} {
		if !strings.Contains(text, "`"+metric+"`") {
			t.Fatalf("metrics contract does not document %q", metric)
		}
	}
}

func TestGoalRuntimeSchemasAreValidJSON(t *testing.T) {
	for _, path := range []string{
		"goalcli-report.schema.json",
		"issue-registry.schema.json",
		"command-registry.schema.json",
		"execution-context.schema.json",
		"conformance-attestation.schema.json",
		"policy.schema.json",
		"execution-evidence.schema.json",
	} {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			var schema map[string]any
			if err := json.Unmarshal(content, &schema); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			if schema["$schema"] == "" || schema["type"] != "object" {
				t.Fatalf("%s must declare object JSON schema, got %#v", path, schema)
			}
		})
	}
}

func TestExecutionContextContractMatchesGovernanceContexts(t *testing.T) {
	schema := readSchema(t, "execution-context.schema.json")

	expected := sortedStrings("local_write", "local_readonly", "ci_pull_request", "ci_main_verify", "release_verify")
	actual := sortedStrings(schema.Properties["context"].Enum...)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("execution context enum drift:\nactual:   %#v\nexpected: %#v", actual, expected)
	}
	requireFields(t, schema.Required, "context", "root", "gowork")
}

func TestExecutionEvidenceContractRequiredFields(t *testing.T) {
	schema := readSchema(t, "execution-evidence.schema.json")
	requireFields(t, schema.Required,
		"evidence_id",
		"command",
		"cwd",
		"branch",
		"commit",
		"exit_code",
		"timestamp",
		"stdout_sha256",
		"artifact_path",
	)
	confidence := sortedStrings(schema.Properties["confidence"].Enum...)
	expectedConfidence := sortedStrings("high", "medium", "low")
	if !reflect.DeepEqual(confidence, expectedConfidence) {
		t.Fatalf("confidence enum drift:\nactual:   %#v\nexpected: %#v", confidence, expectedConfidence)
	}
	if got := schema.Properties["exit_code"].Type; got != "integer" {
		t.Fatalf("exit_code type = %q, want integer", got)
	}
}

func TestExecutionEvidenceContractMatchesEvidenceManifest(t *testing.T) {
	manifest, err := os.ReadFile("../.agent/evidence-artifacts.yaml")
	if err != nil {
		t.Fatalf("read evidence-artifacts.yaml: %v", err)
	}
	text := string(manifest)
	for _, needle := range []string{
		"contracts/execution-evidence.schema.json",
		"execution_evidence",
		"evidence_id",
		"stdout_sha256",
		"exit_code",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf(".agent/evidence-artifacts.yaml missing required marker %q", needle)
		}
	}
}

func requireSchemaFieldMapsToStructField(t *testing.T, schema objectSchema, structType reflect.Type, schemaField string, structField string, schemaType string) {
	t.Helper()

	property, ok := schema.Properties[schemaField]
	if !ok {
		t.Fatalf("schema missing property %q", schemaField)
	}
	if property.Type != schemaType {
		t.Fatalf("schema property %q type = %q, want %q", schemaField, property.Type, schemaType)
	}
	if _, ok := structType.FieldByName(structField); !ok {
		t.Fatalf("%s missing field %s required by schema property %q", structType.Name(), structField, schemaField)
	}
}

func readSchema(t *testing.T, path string) objectSchema {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var schema objectSchema
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return schema
}

func requireFields(t *testing.T, actual []string, expected ...string) {
	t.Helper()
	fields := make(map[string]bool, len(actual))
	for _, field := range actual {
		fields[field] = true
	}
	for _, field := range expected {
		if !fields[field] {
			t.Fatalf("required fields missing %q from %#v", field, actual)
		}
	}
}

func sortedStrings(values ...string) []string {
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return copied
}
