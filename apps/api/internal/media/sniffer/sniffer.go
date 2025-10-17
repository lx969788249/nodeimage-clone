package sniffer

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
)

type MediaType string

const (
	TypeJPEG MediaType = "jpeg"
	TypePNG  MediaType = "png"
	TypeGIF  MediaType = "gif"
	TypeWEBP MediaType = "webp"
	TypeAVIF MediaType = "avif"
	TypeSVG  MediaType = "svg"
)

var ErrUnknownType = errors.New("unknown media type")

type Result struct {
	Type MediaType
	MIME string
}

func Detect(r io.Reader) (Result, []byte, error) {
	head := make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return Result{}, nil, err
	}
	head = head[:n]

	result, err := DetectHead(head)
	return result, head, err
}

func DetectHead(head []byte) (Result, error) {
	if len(head) == 0 {
		return Result{}, ErrUnknownType
	}

	if isJPEG(head) {
		return Result{Type: TypeJPEG, MIME: "image/jpeg"}, nil
	}
	if isPNG(head) {
		return Result{Type: TypePNG, MIME: "image/png"}, nil
	}
	if isGIF(head) {
		return Result{Type: TypeGIF, MIME: "image/gif"}, nil
	}
	if isWEBP(head) {
		return Result{Type: TypeWEBP, MIME: "image/webp"}, nil
	}
	if isAVIF(head) {
		return Result{Type: TypeAVIF, MIME: "image/avif"}, nil
	}
	if isSVG(head) {
		return Result{Type: TypeSVG, MIME: "image/svg+xml"}, nil
	}

	return Result{}, ErrUnknownType
}

func isJPEG(head []byte) bool {
	return len(head) > 3 &&
		head[0] == 0xff &&
		head[1] == 0xd8 &&
		head[2] == 0xff
}

func isPNG(head []byte) bool {
	pngMagic := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	return len(head) >= len(pngMagic) && bytes.Equal(head[:len(pngMagic)], pngMagic)
}

func isGIF(head []byte) bool {
	return len(head) >= 6 && (bytes.Equal(head[:6], []byte("GIF87a")) || bytes.Equal(head[:6], []byte("GIF89a")))
}

func isWEBP(head []byte) bool {
	return len(head) >= 12 &&
		bytes.Equal(head[:4], []byte("RIFF")) &&
		bytes.Equal(head[8:12], []byte("WEBP"))
}

func isAVIF(head []byte) bool {
	if len(head) < 12 {
		return false
	}
	boxType := string(head[8:12])
	return boxType == "ftyp" && bytes.Contains(head[12:], []byte("avif"))
}

func isSVG(head []byte) bool {
	trimmed := strings.TrimSpace(string(head))
	return strings.HasPrefix(trimmed, "<svg") || strings.HasPrefix(trimmed, "<?xml")
}

func MimeTypeFromHTTP(header http.Header) string {
	contentType := header.Get("Content-Type")
	if contentType == "" {
		return ""
	}
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		return strings.TrimSpace(contentType[:idx])
	}
	return strings.TrimSpace(contentType)
}
