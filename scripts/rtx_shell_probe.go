//go:build ignore
// +build ignore

// rtx_shell_probe.go — read-only interactive-shell probe for an RTX router.
//
// The RTX console rejects SSH "exec" requests, so every command has to go
// through an interactive shell session. This mirrors what internal/client does,
// minus the provider plumbing, so a human (or an agent) can ask the live router
// what it actually stores.
//
// Usage:
//
//	RTX_HOST=192.168.1.253 RTX_USERNAME=shin1ohno RTX_KEY=~/.ssh/rtx_hnd_ed25519 \
//	  go run scripts/rtx_shell_probe.go "show environment" "show config | grep schedule"
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	host := os.Getenv("RTX_HOST")
	user := os.Getenv("RTX_USERNAME")
	keyPath := os.Getenv("RTX_KEY")
	password := os.Getenv("RTX_PASSWORD")
	adminPassword := os.Getenv("RTX_ADMIN_PASSWORD")

	if host == "" || user == "" {
		fmt.Fprintln(os.Stderr, "set RTX_HOST and RTX_USERNAME")
		os.Exit(2)
	}

	var auths []ssh.AuthMethod
	if keyPath != "" {
		raw, err := os.ReadFile(keyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read key: %v\n", err)
			os.Exit(1)
		}
		signer, err := ssh.ParsePrivateKey(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse key: %v\n", err)
			os.Exit(1)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}
	if password != "" {
		auths = append(auths, ssh.Password(password))
		auths = append(auths, ssh.KeyboardInteractive(func(_, _ string, qs []string, _ []bool) ([]string, error) {
			answers := make([]string, len(qs))
			for i := range qs {
				answers[i] = password
			}
			return answers, nil
		}))
	}

	conn, err := ssh.Dial("tcp", host+":22", &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         20 * time.Second,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "session: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	if err := session.RequestPty("vt100", 200, 500, ssh.TerminalModes{ssh.ECHO: 0}); err != nil {
		fmt.Fprintf(os.Stderr, "pty: %v\n", err)
		os.Exit(1)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "stdin: %v\n", err)
		os.Exit(1)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "stdout: %v\n", err)
		os.Exit(1)
	}
	session.Stderr = os.Stderr

	out := make(chan string, 1024)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				out <- string(buf[:n])
			}
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "read: %v\n", err)
				}
				close(out)
				return
			}
		}
	}()

	if err := session.Shell(); err != nil {
		fmt.Fprintf(os.Stderr, "shell: %v\n", err)
		os.Exit(1)
	}

	// drain returns everything the router emitted until it went quiet for `idle`.
	drain := func(idle time.Duration) string {
		var sb strings.Builder
		timer := time.NewTimer(idle)
		defer timer.Stop()
		for {
			select {
			case chunk, ok := <-out:
				if !ok {
					return sb.String()
				}
				sb.WriteString(chunk)
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(idle)
			case <-timer.C:
				return sb.String()
			}
		}
	}

	banner := drain(3 * time.Second)
	fmt.Printf("--- banner ---\n%s\n", banner)

	send := func(cmd string) string {
		fmt.Fprintf(stdin, "%s\r", cmd)
		return drain(2500 * time.Millisecond)
	}

	if adminPassword != "" {
		fmt.Printf("--- administrator ---\n%s\n", send("administrator"))
		fmt.Printf("--- (password) ---\n%s\n", send(adminPassword))
	}

	for _, cmd := range os.Args[1:] {
		fmt.Printf("--- %s ---\n%s\n", cmd, send(cmd))
	}

	fmt.Fprint(stdin, "exit\r")
	drain(1500 * time.Millisecond)
}
