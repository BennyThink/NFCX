package attack

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BennyThink/NFCX/internal/nfc"
	runtimebundle "github.com/BennyThink/NFCX/internal/runtime"
)

func TestMFoCEngineAvailabilityAndSuccessfulValidatedDump(t *testing.T) {
	card := nfc.CardInfo{UID: []byte{1, 2, 3, 4}, ATQA: [2]byte{4, 0}, SAK: 0x08}
	key := nfc.Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	engine := NewMFoCEngine(runtimebundle.NewLocator(fixtureRuntimeRoot), MFoCInvocation{
		Card: card, KnownKeys: []MFoCKnownKey{{Sector: 3, Type: nfc.KeyTypeB, Value: key}},
	})
	if err := engine.Available(context.Background()); err != nil {
		t.Fatalf("available: %v", err)
	}
	var visible strings.Builder
	result, err := engine.Run(context.Background(), AttackRequest{
		Device: nfc.DeviceInfo{ConnString: "pn532_uart:/dev/reader with spaces"}, WorkingDir: t.TempDir(),
	}, func(event AttackEvent) { visible.WriteString(event.Data) })
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !result.Validated || result.ExitCode != 0 || result.Version != MFoCVersion || len(result.Artifacts) != 1 || len(result.Artifacts[0].Data) != 1024 {
		t.Fatalf("result = %+v", result)
	}
	if !strings.Contains(strings.ToLower(visible.String()), "ffffffffffff") {
		t.Fatalf("visible log did not retain the MFOC key output: %q", visible.String())
	}
}

func TestMFoCEngineFailureModes(t *testing.T) {
	card := nfc.CardInfo{UID: []byte{1, 2, 3, 4}, SAK: 0x08}
	key := nfc.Key{1, 2, 3, 4, 5, 6}
	for _, test := range []struct {
		name string
		mode string
		want error
	}{{"nonzero", "nonzero", ErrExit}, {"missing", "missing", ErrValidation}, {"corrupt", "corrupt", ErrValidation}, {"dirty-padding", "dirty-padding", ErrValidation}} {
		t.Run(test.name, func(t *testing.T) {
			engine := NewMFoCEngine(runtimebundle.NewLocator(fixtureRuntimeRoot), MFoCInvocation{Card: card, KnownKeys: []MFoCKnownKey{{Sector: 0, Type: nfc.KeyTypeA, Value: key}}})
			engine.extraEnv = map[string]string{"NFCX_FAKE_MFOC_MODE": test.mode}
			_, err := engine.Run(context.Background(), AttackRequest{Device: nfc.DeviceInfo{ConnString: "pn532_uart:/dev/test"}, WorkingDir: t.TempDir()}, nil)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v; want %v", err, test.want)
			}
		})
	}
}

func TestMFoCEngineTimeout(t *testing.T) {
	card := nfc.CardInfo{UID: []byte{1, 2, 3, 4}, SAK: 0x08}
	engine := NewMFoCEngine(runtimebundle.NewLocator(fixtureRuntimeRoot), MFoCInvocation{Card: card, KnownKeys: []MFoCKnownKey{{Sector: 0, Type: nfc.KeyTypeA, Value: nfc.Key{1}}}})
	engine.extraEnv = map[string]string{"NFCX_FAKE_MFOC_MODE": "hang"}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := engine.Run(ctx, AttackRequest{Device: nfc.DeviceInfo{ConnString: "pn532_uart:/dev/test"}, WorkingDir: t.TempDir()}, nil)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("error = %v", err)
	}
}

func TestMFoCArgumentsDeduplicateTypedSeeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "result with spaces.mfd")
	first := nfc.Key{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	second := nfc.Key{1, 2, 3, 4, 5, 6}
	arguments := mfocArguments([]MFoCKnownKey{
		{Sector: 1, Type: nfc.KeyTypeA, Value: first},
		{Sector: 9, Type: nfc.KeyTypeB, Value: first},
		{Sector: 2, Type: nfc.KeyTypeB, Value: second},
	}, path, MFoCModeBalanced)
	want := []string{"-k", "010203040506", "-k", "AABBCCDDEEFF", "-O", path}
	if strings.Join(arguments, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("arguments = %#v; want %#v", arguments, want)
	}
}

func TestMFoCProgressParserAcceptsArbitraryChunks(t *testing.T) {
	var events []AttackEvent
	stream := newMFoCStreamEmitter(func(event AttackEvent) { events = append(events, event) }, 16, "mfoc")
	for _, chunk := range []string{"Sec", "tor: 7, type A", ", probe 0\n[Key: aabb", "ccddeeff]\n"} {
		stream.emit(AttackEvent{Engine: "mfoc", State: StateRunning, Stream: StreamStdout, Data: chunk})
	}
	stream.flush()
	combined := ""
	progress := false
	for _, event := range events {
		combined += event.Data
		progress = progress || strings.Contains(event.Message, "sector 7")
	}
	if !progress || !strings.Contains(strings.ToLower(combined), "aabbccddeeff") {
		t.Fatalf("events = %+v", events)
	}
}

func TestMFoCModeValidation(t *testing.T) {
	for _, test := range []struct {
		mode MFoCMode
		flag string
		ok   bool
	}{
		{MFoCModeBalanced, "", true},
		{MFoCModeHighPerformance, "-H", true},
		{MFoCModeX64, "-X", true},
		{MFoCMode("turbo"), "", false},
	} {
		if test.mode.Valid() != test.ok {
			t.Fatalf("mode %q Valid() = %v; want %v", test.mode, test.mode.Valid(), test.ok)
		}
		if test.mode.flag() != test.flag {
			t.Fatalf("mode %q flag = %q; want %q", test.mode, test.mode.flag(), test.flag)
		}
	}
	invocation := MFoCInvocation{
		Card:      nfc.CardInfo{UID: []byte{1, 2, 3, 4}, SAK: 0x08},
		KnownKeys: []MFoCKnownKey{{Sector: 0, Type: nfc.KeyTypeA, Value: nfc.Key{1}}},
		Mode:      MFoCMode("turbo"),
	}
	if err := validateMFoCInvocation(invocation); err == nil {
		t.Fatal("validateMFoCInvocation accepted an unknown mode")
	}
	invocation.Mode = ""
	if err := validateMFoCInvocation(invocation); err != nil {
		t.Fatalf("empty mode must default to balanced: %v", err)
	}
}

func TestMFoCArgumentsAppendModeFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "result.mfd")
	known := []MFoCKnownKey{{Sector: 0, Type: nfc.KeyTypeA, Value: nfc.Key{1, 2, 3, 4, 5, 6}}}
	for _, test := range []struct {
		mode MFoCMode
		want []string
	}{
		{MFoCModeBalanced, []string{"-k", "010203040506", "-O", path}},
		{MFoCModeHighPerformance, []string{"-k", "010203040506", "-H", "-O", path}},
		{MFoCModeX64, []string{"-k", "010203040506", "-X", "-O", path}},
	} {
		got := mfocArguments(known, path, test.mode)
		if strings.Join(got, "\x00") != strings.Join(test.want, "\x00") {
			t.Fatalf("mode %q arguments = %#v; want %#v", test.mode, got, test.want)
		}
	}
}

func TestMFoCEngineAcceptsAdvertisedModes(t *testing.T) {
	card := nfc.CardInfo{UID: []byte{1, 2, 3, 4}, SAK: 0x08}
	key := nfc.Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	for _, mode := range []MFoCMode{MFoCModeBalanced, MFoCModeHighPerformance, MFoCModeX64} {
		engine := NewMFoCEngine(runtimebundle.NewLocator(fixtureRuntimeRoot), MFoCInvocation{
			Card: card, KnownKeys: []MFoCKnownKey{{Sector: 3, Type: nfc.KeyTypeB, Value: key}}, Mode: mode,
		})
		if err := engine.Available(context.Background()); err != nil {
			t.Fatalf("mode %q availability: %v", mode, err)
		}
	}
}

func TestMFoCHelpWithoutModesIsRejected(t *testing.T) {
	legacy := []byte("Usage: mfoc [-h] [-k key]... [-O output]\n")
	if mfocHelpSupportsMode("-H", legacy) {
		t.Fatal("legacy help text must not advertise -H")
	}
	current := []byte("Usage: mfoc [-h] [-k key]... [-O output] [-H] [-X]\n")
	for _, flag := range []string{"-H", "-X"} {
		if !mfocHelpSupportsMode(flag, current) {
			t.Fatalf("current help text does not advertise %s", flag)
		}
	}
}
