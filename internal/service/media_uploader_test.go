package service

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/stretchr/testify/require"
)

func TestDetectImageByMagic(t *testing.T) {
	t.Parallel()

	jpeg := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46}
	mime, ext, err := detectImage(jpeg)
	require.NoError(t, err)
	require.Equal(t, "image/jpeg", mime)
	require.Equal(t, ".jpg", ext)

	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00}
	mime, ext, err = detectImage(png)
	require.NoError(t, err)
	require.Equal(t, "image/png", mime)
	require.Equal(t, ".png", ext)

	webp := []byte("RIFF....WEBPVP8 ")
	mime, ext, err = detectImage(webp)
	require.NoError(t, err)
	require.Equal(t, "image/webp", mime)
	require.Equal(t, ".webp", ext)

	_, _, err = detectImage([]byte("<html>not an image</html>"))
	require.ErrorIs(t, err, errcode.MediaInvalidType)
}
