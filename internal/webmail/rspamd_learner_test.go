package webmail

import "testing"

func TestRspamdLearnArgs(t *testing.T) {
	tests := []struct {
		verdict string
		want    []string
	}{
		{verdict: "spam", want: []string{"learn_spam", "/tmp/message"}},
		{verdict: "ham", want: []string{"learn_ham", "/tmp/message"}},
	}

	for _, tt := range tests {
		got, err := rspamdLearnArgs("/tmp/message", tt.verdict)
		if err != nil {
			t.Fatalf("rspamdLearnArgs(%q) returned error: %v", tt.verdict, err)
		}
		if len(got) != len(tt.want) {
			t.Fatalf("len(args) = %d, want %d", len(got), len(tt.want))
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("arg[%d] = %q, want %q", i, got[i], tt.want[i])
			}
		}
	}
}

func TestRspamdLearnArgsRejectsUnknownVerdict(t *testing.T) {
	if _, err := rspamdLearnArgs("/tmp/message", "policy"); err == nil {
		t.Fatal("expected unsupported verdict to fail")
	}
}
