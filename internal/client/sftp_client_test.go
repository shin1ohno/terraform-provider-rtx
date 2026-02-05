package client

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSFTPClient is a mock implementation of sftpClientInterface for testing
type mockSFTPClient struct {
	openFunc    func(path string) (sftpFileInterface, error)
	readDirFunc func(path string) ([]os.FileInfo, error)
	createFunc  func(path string) (sftpFileWriteCloserInterface, error)
	closeFunc   func() error
}

func (m *mockSFTPClient) Open(path string) (sftpFileInterface, error) {
	if m.openFunc != nil {
		return m.openFunc(path)
	}
	return nil, errors.New("Open not implemented")
}

func (m *mockSFTPClient) ReadDir(path string) ([]os.FileInfo, error) {
	if m.readDirFunc != nil {
		return m.readDirFunc(path)
	}
	return []os.FileInfo{}, nil
}

func (m *mockSFTPClient) Create(path string) (sftpFileWriteCloserInterface, error) {
	if m.createFunc != nil {
		return m.createFunc(path)
	}
	return nil, errors.New("Create not implemented")
}

func (m *mockSFTPClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// mockSFTPWriteFile is a mock implementation of sftpFileWriteCloserInterface for testing
type mockSFTPWriteFile struct {
	writeFunc func(p []byte) (n int, err error)
	closeFunc func() error
	written   []byte
}

func (m *mockSFTPWriteFile) Write(p []byte) (n int, err error) {
	if m.writeFunc != nil {
		return m.writeFunc(p)
	}
	// Default behavior: store written content
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *mockSFTPWriteFile) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// mockSFTPFile is a mock implementation of sftpFileInterface for testing
type mockSFTPFile struct {
	readFunc  func(p []byte) (n int, err error)
	closeFunc func() error
	content   []byte
	offset    int
}

func (m *mockSFTPFile) Read(p []byte) (n int, err error) {
	if m.readFunc != nil {
		return m.readFunc(p)
	}
	// Default behavior: read from content
	if m.offset >= len(m.content) {
		return 0, io.EOF
	}
	n = copy(p, m.content[m.offset:])
	m.offset += n
	return n, nil
}

func (m *mockSFTPFile) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestSFTPClient_Download_Success(t *testing.T) {
	expectedContent := []byte("# RTX config content\nip route default gateway pp 1\n")

	mockFile := &mockSFTPFile{
		content: expectedContent,
	}

	mockSFTP := &mockSFTPClient{
		openFunc: func(path string) (sftpFileInterface, error) {
			assert.Equal(t, "/system/config0", path)
			return mockFile, nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx := context.Background()
	content, err := client.Download(ctx, "/system/config0")

	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestSFTPClient_Download_FileOpenError(t *testing.T) {
	openError := errors.New("file not found")

	mockSFTP := &mockSFTPClient{
		openFunc: func(path string) (sftpFileInterface, error) {
			return nil, openError
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx := context.Background()
	_, err := client.Download(ctx, "/system/config0")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestSFTPClient_Download_ReadError(t *testing.T) {
	readError := errors.New("read error")

	mockFile := &mockSFTPFile{
		readFunc: func(p []byte) (n int, err error) {
			return 0, readError
		},
	}

	mockSFTP := &mockSFTPClient{
		openFunc: func(path string) (sftpFileInterface, error) {
			return mockFile, nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx := context.Background()
	_, err := client.Download(ctx, "/system/config0")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestSFTPClient_Download_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockSFTP := &mockSFTPClient{
		openFunc: func(path string) (sftpFileInterface, error) {
			return &mockSFTPFile{content: []byte("test")}, nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	_, err := client.Download(ctx, "/system/config0")

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestSFTPClient_Close(t *testing.T) {
	closeCalled := false

	mockSFTP := &mockSFTPClient{
		closeFunc: func() error {
			closeCalled = true
			return nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	err := client.Close()

	require.NoError(t, err)
	assert.True(t, closeCalled)
}

func TestSFTPClient_Close_AlreadyClosed(t *testing.T) {
	closeCount := 0

	mockSFTP := &mockSFTPClient{
		closeFunc: func() error {
			closeCount++
			return nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	// First close
	err := client.Close()
	require.NoError(t, err)

	// Second close should be safe (idempotent)
	err = client.Close()
	require.NoError(t, err)

	// Should only close once
	assert.Equal(t, 1, closeCount)
}

func TestSFTPClient_Download_AfterClose(t *testing.T) {
	mockSFTP := &mockSFTPClient{
		openFunc: func(path string) (sftpFileInterface, error) {
			return &mockSFTPFile{content: []byte("test")}, nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	// Close the client first
	err := client.Close()
	require.NoError(t, err)

	// Try to download after close
	ctx := context.Background()
	_, err = client.Download(ctx, "/system/config0")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSFTPClosed))
}

func TestSFTPClient_WriteFile_Success(t *testing.T) {
	expectedPath := "/ssh/authorized_keys/testuser"
	expectedContent := []byte("ssh-ed25519 AAAAC3... user@host\n")

	mockFile := &mockSFTPWriteFile{}
	var writtenPath string

	mockSFTP := &mockSFTPClient{
		createFunc: func(path string) (sftpFileWriteCloserInterface, error) {
			writtenPath = path
			return mockFile, nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx := context.Background()
	err := client.WriteFile(ctx, expectedPath, expectedContent)

	require.NoError(t, err)
	assert.Equal(t, expectedPath, writtenPath)
	assert.Equal(t, expectedContent, mockFile.written)
}

func TestSFTPClient_WriteFile_CreateError(t *testing.T) {
	createError := errors.New("permission denied")

	mockSFTP := &mockSFTPClient{
		createFunc: func(path string) (sftpFileWriteCloserInterface, error) {
			return nil, createError
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx := context.Background()
	err := client.WriteFile(ctx, "/ssh/authorized_keys/testuser", []byte("key"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")
}

func TestSFTPClient_WriteFile_WriteError(t *testing.T) {
	writeError := errors.New("disk full")

	mockFile := &mockSFTPWriteFile{
		writeFunc: func(p []byte) (int, error) {
			return 0, writeError
		},
	}

	mockSFTP := &mockSFTPClient{
		createFunc: func(path string) (sftpFileWriteCloserInterface, error) {
			return mockFile, nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx := context.Background()
	err := client.WriteFile(ctx, "/ssh/authorized_keys/testuser", []byte("key content"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to file")
}

func TestSFTPClient_WriteFile_PartialWrite(t *testing.T) {
	content := []byte("this is a long key content")

	mockFile := &mockSFTPWriteFile{
		writeFunc: func(p []byte) (int, error) {
			// Only write half the bytes
			return len(p) / 2, nil
		},
	}

	mockSFTP := &mockSFTPClient{
		createFunc: func(path string) (sftpFileWriteCloserInterface, error) {
			return mockFile, nil
		},
	}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx := context.Background()
	err := client.WriteFile(ctx, "/ssh/authorized_keys/testuser", content)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "partial write")
}

func TestSFTPClient_WriteFile_ContextCanceled(t *testing.T) {
	mockSFTP := &mockSFTPClient{}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.WriteFile(ctx, "/ssh/authorized_keys/testuser", []byte("key"))

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestSFTPClient_WriteFile_AfterClose(t *testing.T) {
	mockSFTP := &mockSFTPClient{}

	client := &sftpClientImpl{
		sftpClient: mockSFTP,
	}

	// Close the client first
	err := client.Close()
	require.NoError(t, err)

	// Try to write after close
	ctx := context.Background()
	err = client.WriteFile(ctx, "/ssh/authorized_keys/testuser", []byte("key"))

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSFTPClosed))
}

func TestNewSFTPClient_ConnectionError(t *testing.T) {
	config := &Config{
		Host:     "nonexistent.example.com",
		Port:     22,
		Username: "testuser",
		Password: "testpass",
		Timeout:  1,
	}

	ctx := context.Background()
	_, err := NewSFTPClient(ctx, config)

	// Should fail to connect (connection error is expected)
	require.Error(t, err)
}
