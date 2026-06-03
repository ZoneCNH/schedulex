package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func setupRetroFixture(t *testing.T, withPatch bool) string {
	t.Helper()
	root := t.TempDir()
	must := func(rel, body string) {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	must(".agent/retrospective.md", "# 复盘\n## 失败项\n- foo\n## 补丁\n- bar\n")
	must(".agent/retrospective-template.md",
		"# Retrospective Template\n## Failure\n## Root Cause\n## Patch\n### Prompt Patch\n### Harness Patch\n### Rule Patch\n")
	if withPatch {
		must(".agent/harness-patches.yaml",
			"schema: harness-patches\nentries:\n  - patch_id: PATCH-TEST-001\n    status: PROPOSED\n")
	} else {
		must(".agent/harness-patches.yaml", "schema: harness-patches\nentries: []\n")
	}
	must(".agent/prompt-patches.yaml", "schema: prompt-patches\nentries: []\n")
	must(".agent/rule-patches.yaml", "schema: rule-patches\nentries: []\n")
	must(".agent/harness/gates/retro-gate.yaml", "gate_id: GATE-RETRO\n")
	return root
}

func TestSelfImprovingCheck_Lenient_Passes(t *testing.T) {
	root := setupRetroFixture(t, false)
	var out, errBuf bytes.Buffer
	code := runSelfImprovingCheck("self-improving-check", []string{"--root", root}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("expected 0, got %d; stdout=%s", code, out.String())
	}
}

func TestSelfImprovingCheck_Strict_FailsWithoutEntries(t *testing.T) {
	root := setupRetroFixture(t, false)
	var out, errBuf bytes.Buffer
	code := runSelfImprovingCheck("retro-check", []string{"--root", root, "--strict"}, &out, &errBuf)
	if code != 1 {
		t.Fatalf("expected 1, got %d; stdout=%s", code, out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("RULE-SI-001")) {
		t.Fatalf("expected RULE-SI-001 in output: %s", out.String())
	}
}

func TestSelfImprovingCheck_Strict_PassesWithEntry(t *testing.T) {
	root := setupRetroFixture(t, true)
	var out, errBuf bytes.Buffer
	code := runSelfImprovingCheck("self-improving-check", []string{"--root", root, "--strict"}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("expected 0, got %d; stdout=%s", code, out.String())
	}
}

func TestSelfImprovingCheck_MissingFile_Fails(t *testing.T) {
	root := setupRetroFixture(t, false)
	if err := os.Remove(filepath.Join(root, ".agent/retrospective.md")); err != nil {
		t.Fatalf("rm: %v", err)
	}
	var out, errBuf bytes.Buffer
	code := runSelfImprovingCheck("self-improving-check", []string{"--root", root}, &out, &errBuf)
	if code != 1 {
		t.Fatalf("expected 1, got %d; stdout=%s", code, out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("RULE-RETRO-001")) {
		t.Fatalf("expected RULE-RETRO-001 in output: %s", out.String())
	}
}

func TestSelfImprovingCheck_BadStatus_Fails(t *testing.T) {
	root := setupRetroFixture(t, false)
	bad := "schema: harness-patches\nentries:\n  - patch_id: PATCH-X\n    status: INVALID_X\n"
	if err := os.WriteFile(filepath.Join(root, ".agent/harness-patches.yaml"), []byte(bad), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var out, errBuf bytes.Buffer
	code := runSelfImprovingCheck("self-improving-check", []string{"--root", root}, &out, &errBuf)
	if code != 1 {
		t.Fatalf("expected 1, got %d; stdout=%s", code, out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("RULE-SI-002")) {
		t.Fatalf("expected RULE-SI-002 in output: %s", out.String())
	}
}
