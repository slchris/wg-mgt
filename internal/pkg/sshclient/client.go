package sshclient

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client.
type Client struct {
	host    string
	port    int
	user    string
	keyData string
	timeout time.Duration
}

// NewClient creates a new SSH client.
func NewClient(host string, port int, user, keyData string) *Client {
	return &Client{
		host:    host,
		port:    port,
		user:    user,
		keyData: keyData,
		timeout: 30 * time.Second,
	}
}

// connect establishes an SSH connection.
func (c *Client) connect() (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(c.keyData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// lgtm[go/insecure-hostkeycallback]
		// #nosec G106 - HostKeyCallback is InsecureIgnoreHostKey for internal network use
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.timeout,
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return client, nil
}

// RunCommand executes a command on the remote host.
func (c *Client) RunCommand(cmd string) (string, error) {
	client, err := c.connect()
	if err != nil {
		return "", err
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer func() { _ = session.Close() }()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// WriteFile writes content to a file on the remote host.
func (c *Client) WriteFile(path, content string) error {
	escapedContent := escapeForShell(content)
	cmd := fmt.Sprintf("echo '%s' | sudo tee %s > /dev/null", escapedContent, path)

	_, err := c.RunCommand(cmd)
	return err
}

// escapeForShell escapes content for use in shell commands.
func escapeForShell(s string) string {
	result := bytes.ReplaceAll([]byte(s), []byte("'"), []byte("'\\''"))
	return string(result)
}
