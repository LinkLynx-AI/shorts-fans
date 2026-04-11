package media

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuildShortDeliveryObjectKeys(t *testing.T) {
	t.Parallel()

	keys, err := BuildShortDeliveryObjectKeys(mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"))
	if err != nil {
		t.Fatalf("BuildShortDeliveryObjectKeys() error = %v, want nil", err)
	}

	if got, want := keys.Playback, "shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4"; got != want {
		t.Fatalf("BuildShortDeliveryObjectKeys() playback got %q want %q", got, want)
	}
	if got, want := keys.Poster, "shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster.jpg"; got != want {
		t.Fatalf("BuildShortDeliveryObjectKeys() poster got %q want %q", got, want)
	}
	if got, want := keys.PosterTempBase, "shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster-temp"; got != want {
		t.Fatalf("BuildShortDeliveryObjectKeys() poster temp got %q want %q", got, want)
	}
}

func TestBuildShortDeliveryObjectKeysRejectsNilID(t *testing.T) {
	t.Parallel()

	if _, err := BuildShortDeliveryObjectKeys(uuid.Nil); err == nil {
		t.Fatal("BuildShortDeliveryObjectKeys() error = nil, want error")
	}
}

func TestBuildMainDeliveryObjectKeys(t *testing.T) {
	t.Parallel()

	keys, err := BuildMainDeliveryObjectKeys(mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"))
	if err != nil {
		t.Fatalf("BuildMainDeliveryObjectKeys() error = %v, want nil", err)
	}

	if got, want := keys.Playback, "mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4"; got != want {
		t.Fatalf("BuildMainDeliveryObjectKeys() playback got %q want %q", got, want)
	}
	if got, want := keys.Poster, "mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster.jpg"; got != want {
		t.Fatalf("BuildMainDeliveryObjectKeys() poster got %q want %q", got, want)
	}
	if got, want := keys.PosterTempBase, "mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster-temp"; got != want {
		t.Fatalf("BuildMainDeliveryObjectKeys() poster temp got %q want %q", got, want)
	}
}

func TestBuildMainDeliveryObjectKeysRejectsNilID(t *testing.T) {
	t.Parallel()

	if _, err := BuildMainDeliveryObjectKeys(uuid.Nil); err == nil {
		t.Fatal("BuildMainDeliveryObjectKeys() error = nil, want error")
	}
}
