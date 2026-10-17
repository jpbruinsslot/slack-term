package token

import (
	"net/url"
	"strings"
	"testing"
)

func TestDefaultConstants(t *testing.T) {
	if !strings.Contains(DefaultRedirectURI, "jpbruinsslot.github.io/slack-term-auth/") {
		t.Errorf("expected DefaultRedirectURI to point to jpbruinsslot, got %s", DefaultRedirectURI)
	}

	if strings.Contains(DefaultRedirectURI, "erroneousboat") {
		t.Errorf("DefaultRedirectURI must not contain erroneousboat: %s", DefaultRedirectURI)
	}

	if DefaultClientID == "" {
		t.Error("DefaultClientID must not be empty")
	}

	if DefaultClientSecret == "" {
		t.Error("DefaultClientSecret must not be empty")
	}

	if DefaultScope != "client" {
		t.Errorf("expected DefaultScope to be 'client', got '%s'", DefaultScope)
	}
}

func TestMethod1AuthorizeURL_Formulation(t *testing.T) {
	rawURL := Method1AuthorizeURL()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse generated authorize URL: %v", err)
	}

	if parsed.Scheme != "https" {
		t.Errorf("expected scheme https, got %s", parsed.Scheme)
	}

	if parsed.Host != "slack.com" {
		t.Errorf("expected host slack.com, got %s", parsed.Host)
	}

	if parsed.Path != "/oauth/authorize" {
		t.Errorf("expected path /oauth/authorize, got %s", parsed.Path)
	}

	query := parsed.Query()

	if query.Get("client_id") != DefaultClientID {
		t.Errorf("expected client_id %s, got %s", DefaultClientID, query.Get("client_id"))
	}

	if query.Get("scope") != DefaultScope {
		t.Errorf("expected scope %s, got %s", DefaultScope, query.Get("scope"))
	}

	expectedRedirect := "https://jpbruinsslot.github.io/slack-term-auth/"
	if query.Get("redirect_uri") != expectedRedirect {
		t.Errorf("expected redirect_uri %s, got %s", expectedRedirect, query.Get("redirect_uri"))
	}

	if strings.Contains(rawURL, "erroneousboat") {
		t.Errorf("authorize URL should not contain legacy handle erroneousboat: %s", rawURL)
	}
}

func TestMethod2AuthorizeURL_Formulation(t *testing.T) {
	rawURL := Method2AuthorizeURL()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse Method 2 authorize URL: %v", err)
	}

	query := parsed.Query()
	if query.Get("redirect_uri") != DefaultFakeHost {
		t.Errorf("expected redirect_uri %s, got %s", DefaultFakeHost, query.Get("redirect_uri"))
	}

	if query.Get("client_id") != DefaultClientID {
		t.Errorf("expected client_id %s, got %s", DefaultClientID, query.Get("client_id"))
	}
}

func TestAuthorizeURL_CustomRedirect(t *testing.T) {
	gen := NewGenerator()
	customRedirect := "https://example.com/oauth/callback?env=test"
	rawURL := gen.AuthorizeURL(customRedirect)

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse custom authorize URL: %v", err)
	}

	query := parsed.Query()
	if query.Get("redirect_uri") != customRedirect {
		t.Errorf("expected redirect_uri %s, got %s", customRedirect, query.Get("redirect_uri"))
	}
}

func TestTokenRetrievalURL_Formulation(t *testing.T) {
	code := "4321.8765.testauthcode"
	rawURL := TokenRetrievalURL(code)

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse token access URL: %v", err)
	}

	if parsed.Scheme != "https" {
		t.Errorf("expected scheme https, got %s", parsed.Scheme)
	}

	if parsed.Host != "slack.com" {
		t.Errorf("expected host slack.com, got %s", parsed.Host)
	}

	if parsed.Path != "/api/oauth.access" {
		t.Errorf("expected path /api/oauth.access, got %s", parsed.Path)
	}

	query := parsed.Query()
	if query.Get("client_id") != DefaultClientID {
		t.Errorf("expected client_id %s, got %s", DefaultClientID, query.Get("client_id"))
	}

	if query.Get("client_secret") != DefaultClientSecret {
		t.Errorf("expected client_secret %s, got %s", DefaultClientSecret, query.Get("client_secret"))
	}

	if query.Get("code") != code {
		t.Errorf("expected code %s, got %s", code, query.Get("code"))
	}
}

func TestCurlCommand(t *testing.T) {
	gen := NewGenerator()
	code := "sample-test-code-999"
	cmd := gen.CurlCommand(code)

	if !strings.HasPrefix(cmd, "curl -s \"https://slack.com/api/oauth.access?") {
		t.Errorf("unexpected curl command prefix: %s", cmd)
	}

	if !strings.Contains(cmd, "code=sample-test-code-999") {
		t.Errorf("expected curl command to contain code parameter: %s", cmd)
	}

	if !strings.HasSuffix(cmd, "\"") {
		t.Errorf("expected curl command to end with closing quote: %s", cmd)
	}
}

func TestExtractCode(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "full github pages callback url with state",
			input:       "https://jpbruinsslot.github.io/slack-term-auth/?code=91899392594.3827.abc123xyz&state=ok",
			expected:    "91899392594.3827.abc123xyz",
			expectError: false,
		},
		{
			name:        "fake host url from method 2",
			input:       "http://not.a.host?code=auth_code_987654&state=",
			expected:    "auth_code_987654",
			expectError: false,
		},
		{
			name:        "bare code input without url wrapping",
			input:       "91899392594.3827.abc123xyz",
			expected:    "91899392594.3827.abc123xyz",
			expectError: false,
		},
		{
			name:        "input with surrounding whitespace",
			input:       "  https://jpbruinsslot.github.io/slack-term-auth/?code=trimmed_code   ",
			expected:    "trimmed_code",
			expectError: false,
		},
		{
			name:        "empty string input",
			input:       "   ",
			expected:    "",
			expectError: true,
		},
		{
			name:        "url without code parameter",
			input:       "https://jpbruinsslot.github.io/slack-term-auth/?error=access_denied",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExtractCode(tc.input)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for input %q: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("expected code %q, got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestCustomGenerator(t *testing.T) {
	custom := &Generator{
		ClientID:     "custom-client-id",
		ClientSecret: "custom-secret",
		Scope:        "identity.basic",
		RedirectURI:  "https://custom.example.org/callback",
	}

	authURL := custom.AuthorizeURL("")
	parsedAuth, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse custom generator auth URL: %v", err)
	}

	if parsedAuth.Query().Get("client_id") != "custom-client-id" {
		t.Errorf("expected custom client_id, got %s", parsedAuth.Query().Get("client_id"))
	}
	if parsedAuth.Query().Get("scope") != "identity.basic" {
		t.Errorf("expected custom scope, got %s", parsedAuth.Query().Get("scope"))
	}
	if parsedAuth.Query().Get("redirect_uri") != "https://custom.example.org/callback" {
		t.Errorf("expected custom redirect_uri, got %s", parsedAuth.Query().Get("redirect_uri"))
	}

	accessURL := custom.AccessURL("my-code")
	parsedAccess, err := url.Parse(accessURL)
	if err != nil {
		t.Fatalf("failed to parse custom generator access URL: %v", err)
	}
	if parsedAccess.Query().Get("client_secret") != "custom-secret" {
		t.Errorf("expected custom client_secret, got %s", parsedAccess.Query().Get("client_secret"))
	}
}
