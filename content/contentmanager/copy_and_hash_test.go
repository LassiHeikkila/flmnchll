package main

import (
	"bytes"
	"crypto"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyAndHash(t *testing.T) {
	tests := map[int]struct {
		inputReader    io.Reader
		expectedLength int
		algo           crypto.Hash
		expectedHash   []byte
	}{
		0: {
			inputReader:    bytes.NewBufferString("hello world"),
			expectedLength: 11,
			algo:           crypto.MD5,
			expectedHash:   []byte{0x5e, 0xb6, 0x3b, 0xbb, 0xe0, 0x1e, 0xee, 0xd0, 0x93, 0xcb, 0x22, 0xbb, 0x8f, 0x5a, 0xcd, 0xc3},
		},
		1: {
			inputReader:    bytes.NewBufferString("hello world"),
			expectedLength: 11,
			algo:           crypto.SHA1,
			expectedHash:   []byte{0x2a, 0xae, 0x6c, 0x35, 0xc9, 0x4f, 0xcf, 0xb4, 0x15, 0xdb, 0xe9, 0x5f, 0x40, 0x8b, 0x9c, 0xe9, 0x1e, 0xe8, 0x46, 0xed},
		},
		2: {
			inputReader:    bytes.NewBufferString("hello world"),
			expectedLength: 11,
			algo:           crypto.SHA256,
			expectedHash:   []byte{0xb9, 0x4d, 0x27, 0xb9, 0x93, 0x4d, 0x3e, 0x08, 0xa5, 0x2e, 0x52, 0xd7, 0xda, 0x7d, 0xab, 0xfa, 0xc4, 0x84, 0xef, 0xe3, 0x7a, 0x53, 0x80, 0xee, 0x90, 0x88, 0xf7, 0xac, 0xe2, 0xef, 0xcd, 0xe9},
		},
		3: {
			inputReader:    bytes.NewBufferString(veryLongText),
			expectedLength: 8205,
			algo:           crypto.SHA256,
			expectedHash:   []byte{0xec, 0x9a, 0xbc, 0x71, 0x11, 0x35, 0x9e, 0x53, 0xd9, 0x69, 0x3c, 0x4f, 0xd1, 0xfc, 0x2c, 0x44, 0xa3, 0x05, 0xc4, 0x43, 0x79, 0x54, 0x90, 0x03, 0x1d, 0xa4, 0x4b, 0x33, 0xfc, 0x47, 0xc5, 0x91},
		},
	}

	for n, tc := range tests {
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			dst := bytes.Buffer{}
			n, h, err := CopyAndHash(&dst, tc.inputReader, tc.algo)
			require.Nil(t, err)
			require.Equal(t, tc.expectedLength, n)
			require.Equal(t, tc.expectedHash, h)
		})
	}
}
