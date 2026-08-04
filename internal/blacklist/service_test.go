package blacklist

import "testing"

func TestCacheKeyMatchesJavaPrefixes(t *testing.T) {
	tests := map[string]string{
		"DEVICE_ID":   "blacklist:device:x",
		"USER_ID":     "blacklist:user:x",
		"MERCHANT_ID": "blacklist:merchant:x",
		"IP_ADDRESS":  "blacklist:account:x",
	}
	for typ, want := range tests {
		if got := cacheKey(typ, "x"); got != want {
			t.Fatalf("cacheKey(%s) = %s, want %s", typ, got, want)
		}
	}
}

func TestReasonPriorityMatchesJava(t *testing.T) {
	if got := reasonFor([]string{"SENDER_USER_ID:1"}); got != "BLACKLISTED_SENDER" {
		t.Fatalf("reason = %s", got)
	}
	if got := reasonFor([]string{"DEVICE_ID:d", "SENDER_USER_ID:1"}); got != "BLACKLISTED_DEVICE" {
		t.Fatalf("reason = %s", got)
	}
}
