package pgping

import "testing"

func TestStorageRTTSamplesReturnsEmptySliceForNil(t *testing.T) {
	got := storageRTTSamples(nil)
	if got == nil {
		t.Fatal("expected nil RTT samples to become an empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty RTT samples, got %#v", got)
	}
}

func TestStorageRTTSamplesCopiesValues(t *testing.T) {
	input := []float64{10, 20}
	got := storageRTTSamples(input)
	input[0] = 99

	if len(got) != 2 || got[0] != 10 || got[1] != 20 {
		t.Fatalf("expected copied RTT samples, got %#v", got)
	}
}
