// +build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"
)

var log zerolog.Logger

func init() {
	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).
		With().Timestamp().Logger()
}

func main() {
	host := os.Getenv("RTX_HOST")
	username := os.Getenv("RTX_USERNAME")
	password := os.Getenv("RTX_ADMIN_PASSWORD") // Use admin password for SFTP

	if host == "" || username == "" || password == "" {
		log.Fatal().Msg("Set RTX_HOST, RTX_USERNAME, and RTX_ADMIN_PASSWORD environment variables")
	}

	// SSH config
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to SSH
	addr := fmt.Sprintf("%s:22", host)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect")
	}
	defer conn.Close()

	// Create SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create SFTP client")
	}
	defer client.Close()

	// Download config0
	configPath := "/system/config0"
	remoteFile, err := client.Open(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", configPath).Msg("Failed to open remote file")
	}
	defer remoteFile.Close()

	// Read content
	content, err := io.ReadAll(remoteFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read remote file")
	}

	// Write to local file
	outputPath := "/tmp/rtx_config.txt"
	err = os.WriteFile(outputPath, content, 0600)
	if err != nil {
		log.Fatal().Err(err).Str("path", outputPath).Msg("Failed to write local file")
	}

	fmt.Printf("Downloaded %d bytes to %s\n", len(content), outputPath)

	// Also print the content (with sensitive data warning)
	fmt.Println("\n=== RAW CONFIG (SENSITIVE DATA MAY BE PRESENT) ===")
	fmt.Println(string(content))
}
