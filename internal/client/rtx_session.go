package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// rtxShellSession represents a persistent shell session for RTX routers
type rtxShellSession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	reader  *bufio.Reader
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// newRTXShellSession creates a new persistent shell session
func newRTXShellSession(ctx context.Context, client *ssh.Client) (*rtxShellSession, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Request PTY - RTX routers require this
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Use wide terminal to prevent line wrapping for long filter lists
	// RTX config lines can exceed 200 characters (e.g., secure filter with 13+ IDs)
	// RequestPty parameters: term, height, width, modes
	if err := session.RequestPty("vt100", 40, 512, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := session.Shell(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	// Use context.Background() for the shell session to ensure it stays alive
	// The session lifetime is managed by Close() method, not by context cancellation
	shellCtx, cancel := context.WithCancel(context.Background())
	r := &rtxShellSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		reader:  bufio.NewReader(stdout),
		ctx:     shellCtx,
		cancel:  cancel,
	}

	// Read initial prompt
	initialOutput, err := r.readUntilPrompt()
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("failed to read initial prompt: %w", err)
	}
	log.Printf("[DEBUG] RTX initial prompt: %q", string(initialOutput))

	// Set character encoding to ASCII
	encodingOutput, err := r.executeCommand("console character en.ascii")
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("failed to set character encoding: %w", err)
	}
	log.Printf("[DEBUG] RTX encoding command output: %q", string(encodingOutput))

	return r, nil
}

// Send executes a command and returns the output
func (r *rtxShellSession) Send(cmd string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("[DEBUG] RTX sending command: %q", cmd)
	output, err := r.executeCommand(cmd)
	if err != nil {
		log.Printf("[ERROR] RTX command failed: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] RTX command output: %q", string(output))

	return output, nil
}

// executeCommand sends a command and reads the response
func (r *rtxShellSession) executeCommand(cmd string) ([]byte, error) {
	// Check if session is still alive
	select {
	case <-r.ctx.Done():
		return nil, fmt.Errorf("session context cancelled")
	default:
	}

	// Send command with retry
	for retry := 0; retry < 3; retry++ {
		if retry > 0 {
			log.Printf("[DEBUG] RTX retrying command (attempt %d)", retry+1)
			time.Sleep(100 * time.Millisecond)
		}

		_, err := fmt.Fprintln(r.stdin, cmd)
		if err == nil {
			break
		}
		if err == io.EOF && retry < 2 {
			continue
		}
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response until next prompt
	output, err := r.readUntilPrompt()
	if err != nil {
		return nil, fmt.Errorf("failed to read command output: %w", err)
	}

	// RTX routers echo the command back, so we need to remove it
	// Find the first newline after the command
	cmdEcho := []byte(cmd + "\r\n")
	if bytes.HasPrefix(output, cmdEcho) {
		output = output[len(cmdEcho):]
	} else {
		// Try with just \n
		cmdEcho = []byte(cmd + "\n")
		if bytes.HasPrefix(output, cmdEcho) {
			output = output[len(cmdEcho):]
		}
	}

	// The output includes the prompt at the end, which we keep
	// This is important for the prompt detector to work correctly
	return output, nil
}

// readUntilPrompt reads output until finding a prompt
func (r *rtxShellSession) readUntilPrompt() ([]byte, error) {
	var output bytes.Buffer
	timeout := time.After(10 * time.Second)
	
	for {
		select {
		case <-r.ctx.Done():
			return output.Bytes(), fmt.Errorf("session context cancelled")
		case <-timeout:
			return output.Bytes(), fmt.Errorf("timeout waiting for prompt")
		default:
			// Read one byte at a time
			b, err := r.reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				return output.Bytes(), fmt.Errorf("failed to read output: %w", err)
			}

			// Write to output
			output.WriteByte(b)

			// Check for prompt
			if b == '>' || b == '#' {
				// Peek at the next byte
				if nextBytes, err := r.reader.Peek(1); err == nil {
					next := nextBytes[0]
					if next == ' ' || next == '\n' || next == '\r' {
						// Found prompt, consume the space/newline
						r.reader.ReadByte()
						output.WriteByte(next)
						return output.Bytes(), nil
					}
				} else {
					// Can't peek, assume it's a prompt
					return output.Bytes(), nil
				}
			}
		}
	}
}

// Close closes the shell session
func (r *rtxShellSession) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cancel()

	// Send exit command
	fmt.Fprintln(r.stdin, "exit")
	
	// Close the session
	if r.session != nil {
		return r.session.Close()
	}
	return nil
}