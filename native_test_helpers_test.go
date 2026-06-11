package bid754

import "testing"

func requireNative(t *testing.T) {
	t.Helper()
	if !NativeBackendEnabled() {
		t.Skip("native decimal backends disabled; run with CGO and -tags bid754_native after sourcing .env.sh")
	}
}

func requireNativeBenchmark(b *testing.B) {
	b.Helper()
	if !NativeBackendEnabled() {
		b.Skip("native decimal backends disabled; run with CGO and -tags bid754_native after sourcing .env.sh")
	}
}
