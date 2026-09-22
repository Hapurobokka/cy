package keys

import (
	"testing"

	"github.com/cfoust/cy/pkg/emu"

	"github.com/stretchr/testify/require"
)

// The fields of a Kitty key event are key ; modifiers[:type] ; text. The
// number for the letter form (CSI 1;mods A) is mandatory whenever anything
// follows it.
func TestKittyEncodeFields(t *testing.T) {
	const (
		all   = emu.KeyReportAllKeys
		types = emu.KeyReportEventTypes
		alt   = emu.KeyReportAlternateKeys
		text  = emu.KeyReportAssociatedText
	)

	cases := []struct {
		name     string
		key      Key
		protocol emu.KeyProtocol
		want     string
	}{
		{
			"up press",
			Key{Code: KittyKeyUp},
			all | types,
			"\x1b[A",
		},
		{
			"up repeat",
			Key{Code: KittyKeyUp, Type: KeyEventRepeat},
			all | types,
			"\x1b[1;1:2A",
		},
		{
			"up release",
			Key{Code: KittyKeyUp, Type: KeyEventRelease},
			all | types,
			"\x1b[1;1:3A",
		},
		{
			"ctrl+up",
			Key{Code: KittyKeyUp, Mod: KeyModCtrl},
			all,
			"\x1b[1;5A",
		},
		{
			"f1 press",
			Key{Code: KittyKeyF1},
			all,
			"\x1b[P",
		},
		{
			"f1 repeat",
			Key{Code: KittyKeyF1, Type: KeyEventRepeat},
			all | types,
			"\x1b[1;1:2P",
		},
		{
			"page up",
			Key{Code: KittyKeyPageUp},
			all,
			"\x1b[5~",
		},
		{
			"rune with text",
			Key{Code: 'o', Shifted: 'O', Text: "o"},
			all | text,
			"\x1b[111;;111u",
		},
		{
			"rune with alternate key and text",
			Key{Code: 'o', Shifted: 'O', Text: "o"},
			all | alt | text,
			"\x1b[111:79;;111u",
		},
		{
			"rune release",
			Key{Code: 'o', Shifted: 'O', Text: "o", Type: KeyEventRelease},
			all | types,
			"\x1b[111;1:3u",
		},
		{
			"ctrl+space with text",
			Key{Code: ' ', Mod: KeyModCtrl, Text: " "},
			all | text,
			"\x1b[32;5;32u",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, ok := c.key.Bytes(emu.DefaultMode, c.protocol)
			require.True(t, ok)
			require.Equal(t, c.want, string(data))
		})
	}
}

func TestKittyDecodeFixes(t *testing.T) {
	t.Run("altgr is a modifier", func(t *testing.T) {
		event, _ := Read([]byte("\x1b[57453u"))
		key, ok := event.(Key)
		require.True(t, ok)
		require.Equal(t, rune(KittyIsoLevel3Shift), key.Code)
		require.True(t, isModifierKey(key.Code))

		// A pane that only asked for disambiguation must never see it.
		_, ok = key.Bytes(emu.DefaultMode, emu.KeyDisambiguateEscape)
		require.False(t, ok)
	})

	t.Run("lock modifiers are dropped", func(t *testing.T) {
		event, _ := Read([]byte("\x1b[1;129A")) // up with Num Lock
		key, ok := event.(Key)
		require.True(t, ok)
		require.Equal(t, rune(KittyKeyUp), key.Code)
		require.Equal(t, KeyModifiers(0), key.Mod)

		data, ok := key.Bytes(emu.DefaultMode, emu.KeyLegacy)
		require.True(t, ok)
		require.Equal(t, "\x1b[A", string(data))

		// And they do not change how a binding is matched.
		event, _ = Read([]byte("\x1b[32;133;32u")) // ctrl+space with Num Lock
		key, ok = event.(Key)
		require.True(t, ok)
		require.Equal(t, "ctrl+space", key.String())
	})

	t.Run("the private use area is never text", func(t *testing.T) {
		_, ok := Key{Code: KittyPUAStart}.Bytes(emu.DefaultMode, emu.KeyLegacy)
		require.False(t, ok)
	})
}
