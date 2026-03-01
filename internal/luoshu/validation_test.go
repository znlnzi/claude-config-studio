package luoshu

import "testing"

func TestPreValidateKey_TooShort(t *testing.T) {
	_, err := PreValidateKey("short")
	if err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestPreValidateKey_OpenAI(t *testing.T) {
	provider, err := PreValidateKey("sk-proj-abcdefghijklmnopqrst")
	if err != nil {
		t.Fatal(err)
	}
	if provider != "openai" {
		t.Errorf("expected openai, got %s", provider)
	}
}

func TestPreValidateKey_Anthropic(t *testing.T) {
	provider, err := PreValidateKey("sk-ant-abcdefghijklmnopqrst")
	if err != nil {
		t.Fatal(err)
	}
	if provider != "anthropic" {
		t.Errorf("expected anthropic, got %s", provider)
	}
}

func TestPreValidateKey_AWS(t *testing.T) {
	provider, err := PreValidateKey("AKIAabcdefghijklmnopqrst")
	if err != nil {
		t.Fatal(err)
	}
	if provider != "aws" {
		t.Errorf("expected aws, got %s", provider)
	}
}

func TestPreValidateKey_GitHub(t *testing.T) {
	provider, err := PreValidateKey("ghp_abcdefghijklmnopqrst")
	if err != nil {
		t.Fatal(err)
	}
	if provider != "github" {
		t.Errorf("expected github, got %s", provider)
	}
}

func TestPreValidateKey_GenericSK(t *testing.T) {
	provider, err := PreValidateKey("sk-abcdefghijklmnopqrst")
	if err != nil {
		t.Fatal(err)
	}
	if provider != "openai-compatible" {
		t.Errorf("expected openai-compatible, got %s", provider)
	}
}

func TestPreValidateKey_Unknown(t *testing.T) {
	provider, err := PreValidateKey("someunknownkeyformat12345")
	if err != nil {
		t.Fatal(err)
	}
	if provider != "unknown" {
		t.Errorf("expected unknown, got %s", provider)
	}
}

func TestMaskKey_Normal(t *testing.T) {
	masked := MaskKey("sk-proj-abcdefghijklmnop")
	if masked != "sk-****mnop" {
		t.Errorf("expected 'sk-****mnop', got %q", masked)
	}
}

func TestMaskKey_Short(t *testing.T) {
	masked := MaskKey("short")
	if masked != "****" {
		t.Errorf("expected '****', got %q", masked)
	}
}

func TestMaskKey_Empty(t *testing.T) {
	masked := MaskKey("")
	if masked != "****" {
		t.Errorf("expected '****', got %q", masked)
	}
}

func TestClassifyResponse_OK(t *testing.T) {
	ok, status, err := classifyResponse(200)
	if !ok || status != "ok" || err != nil {
		t.Errorf("expected (true, ok, nil), got (%v, %s, %v)", ok, status, err)
	}
}

func TestClassifyResponse_AuthFailed(t *testing.T) {
	ok, status, _ := classifyResponse(401)
	if ok || status != "auth_failed" {
		t.Errorf("expected (false, auth_failed), got (%v, %s)", ok, status)
	}
	ok, status, _ = classifyResponse(403)
	if ok || status != "auth_failed" {
		t.Errorf("expected (false, auth_failed), got (%v, %s)", ok, status)
	}
}

func TestClassifyResponse_QuotaExceeded(t *testing.T) {
	ok, status, _ := classifyResponse(429)
	if ok || status != "quota_exceeded" {
		t.Errorf("expected (false, quota_exceeded), got (%v, %s)", ok, status)
	}
}

func TestClassifyResponse_ServerError(t *testing.T) {
	ok, status, _ := classifyResponse(500)
	if ok || status != "network_error" {
		t.Errorf("expected (false, network_error), got (%v, %s)", ok, status)
	}
}
