package ref

import "testing"

func TestIsRef(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"cred://work/TOKEN", true},
		{"cred://", true},
		{"cred:/work/TOKEN", true},
		{"cred:work/TOKEN", true},
		{"cred//work/TOKEN", true},
		{"credential", false},
		{"CRED://work/TOKEN", false},
		{"op://vault/item", false},
		{"plain-value", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsRef(tt.value); got != tt.want {
			t.Errorf("IsRef(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestParse_Valid(t *testing.T) {
	r, err := Parse("cred://work/JIRA_TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	if r.Namespace != "work" || r.Key != "JIRA_TOKEN" {
		t.Errorf("got %+v, want work/JIRA_TOKEN", r)
	}
	if r.String() != "cred://work/JIRA_TOKEN" {
		t.Errorf("String() = %q", r.String())
	}
}

func TestParse_Invalid(t *testing.T) {
	invalid := []string{
		"cred://",
		"cred://work",
		"cred://work/",
		"cred:///TOKEN",
		"cred://a/b/c",
		"cred:/work/TOKEN",
		"cred//work/TOKEN",
		"cred://work/TOKEN # jira",
		"cred://wo rk/TOKEN",
		"cred://work/TO$KEN",
		"CRED://work/TOKEN",
		"plain-value",
		"",
	}
	for _, v := range invalid {
		if _, err := Parse(v); err == nil {
			t.Errorf("Parse(%q): expected error", v)
		}
	}
}

func TestParse_AllowedCharacters(t *testing.T) {
	r, err := Parse("cred://my-project.v2/api_key-1.prod")
	if err != nil {
		t.Fatal(err)
	}
	if r.Namespace != "my-project.v2" || r.Key != "api_key-1.prod" {
		t.Errorf("got %+v", r)
	}
}
