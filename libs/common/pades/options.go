package pades

import (
	"os"
	"strings"

	"github.com/digitorus/pdfsign/verify"
)

func verifyOptions() *verify.VerifyOptions {
	opts := verify.DefaultVerifyOptions()
	if allowUntrustedRoots() {
		opts.AllowUntrustedRoots = true
	}
	return opts
}

func allowUntrustedRoots() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("PADES_ALLOW_UNTRUSTED_ROOTS")))
	return v == "1" || v == "true" || v == "yes"
}
