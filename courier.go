package courier

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Courier struct {
	SSHConnection *ssh.Client
}

func NewCourierWithPassword(host string, port int, username, password string, timeout time.Duration) (*Courier, error) {
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout: timeout,
	}

	portString := strconv.Itoa(port)
	sshConnection, err := ssh.Dial("tcp", host+":"+portString, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("Error dialing host: %s on port: %s: %s", host, portString, err)
	}

	courier := &Courier{}
	courier.SSHConnection = sshConnection

	return courier, nil
}

func (c *Courier) Run(cmd string) (string, error) {

	if c.SSHConnection == nil {
		return "", fmt.Errorf("Error getting SSH connection")
	}
	sshSession, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("Error creating new SSH session: %s", err)
	}

	defer sshSession.Close()

	var outBuffer, errBuffer bytes.Buffer
	sshSession.Stdout = &outBuffer
	sshSession.Stderr = &errBuffer

	err = sshSession.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("Error running: %s: %s", cmd, err)
	}

	outString := outBuffer.String()
	errString := errBuffer.String()
	if errString != "" {
		return outString, fmt.Errorf(errString)
	}
	return outString, nil
}

func (c *Courier) RunWithEnv(cmd string, env []string) (string, error) {

	if c.SSHConnection == nil {
		return "", fmt.Errorf("Error getting SSH connection")
	}

	sshSession, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("Error creating new SSH session: %s", err)
	}

	defer sshSession.Close()

	for _, e := range env {
		variable := strings.Split(e, "=")
		if len(variable) != 2 {
			continue
		}

		if err := sshSession.Setenv(variable[0], variable[1]); err != nil {
			return "", fmt.Errorf("Error setting env for cmd: %s", err)
		}
	}

	var outBuffer, errBuffer bytes.Buffer
	sshSession.Stdout = &outBuffer
	sshSession.Stderr = &errBuffer

	err = sshSession.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("Error running: %s: %s", cmd, err)
	}

	outString := outBuffer.String()
	errString := errBuffer.String()
	if errString != "" {
		return outString, fmt.Errorf(errString)
	}
	return outString, nil
}

func (c *Courier) Close() error {

	if c.SSHConnection == nil {
		return fmt.Errorf("Error getting SSH connection")
	}

	err := c.SSHConnection.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Courier) NewSession() (*ssh.Session, error) {
	sshSession, err := c.SSHConnection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Error creating new SSH session: %s", err)
	}

	sshModes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sshSession.RequestPty("xterm", 80, 40, sshModes); err != nil {
		sshSession.Close()
		return nil, fmt.Errorf("Error request for pseudo terminal failed: %s", err)
	}

	return sshSession, nil

}
