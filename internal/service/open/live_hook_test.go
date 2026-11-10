package open

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSRSPlayStreamName(t *testing.T) {
	require.Equal(t, "s1", srsPlayStreamName("s1.m3u8"))
	require.Equal(t, "s1", srsPlayStreamName("s1.flv"))
	require.Equal(t, "S1", srsPlayStreamName("S1.M3U8"))
	require.Equal(t, "s1", srsPlayStreamName(" s1 "))
	require.Equal(t, "s1", srsPlayStreamName("s1"))
	require.Equal(t, "", srsPlayStreamName(""))
}
