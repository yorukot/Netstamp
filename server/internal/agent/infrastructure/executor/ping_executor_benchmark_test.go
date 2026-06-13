package executor

import "testing"

func BenchmarkMakePingPayload(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = makePingPayload(56)
	}
}
