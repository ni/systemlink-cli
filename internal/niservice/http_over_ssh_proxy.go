package niservice

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/elazarl/goproxy"
	"golang.org/x/crypto/ssh"
)

// SSHConfig contains all the parameters needed to open an SSH connection
type SSHConfig struct {
	HostName  string
	KeyFile   string
	KnownHost string
	UserName  string
}

// HTTPOverSSHProxy tunnels HTTP requests through SSH by opening a proxy
// and forwarding all requests
type HTTPOverSSHProxy struct{}

// Start connects through SSH to the given hostname and spins up the HTTP proxy
// which forwards all requests
func (proxy *HTTPOverSSHProxy) Start(sshConfig SSHConfig) (string, error) {
	client, err := proxy.connectToProxy(sshConfig)
	if err != nil {
		return "", err
	}

	httpProxy := goproxy.NewProxyHttpServer()
	httpProxy.ConnectDial = client.Dial

	httpServer := &http.Server{Handler: httpProxy}
	httpServerListener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	go httpServer.Serve(httpServerListener)

	port := httpServerListener.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf("localhost:%v", port), nil
}

func (proxy *HTTPOverSSHProxy) connectToProxy(sshConfig SSHConfig) (*ssh.Client, error) {
	config := proxy.clientConfig(sshConfig)
	client, err := ssh.Dial("tcp", sshConfig.HostName, config)
	if err != nil {
		return nil, fmt.Errorf("Could not SSH into %s. Make sure that you have provided the correct SSH key and the server is a known host. Error: %s", sshConfig.HostName, err)
	}

	return client, nil
}

func (proxy *HTTPOverSSHProxy) clientConfig(sshConfig SSHConfig) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: sshConfig.UserName,
		Auth: []ssh.AuthMethod{
			proxy.publicKeyFile(sshConfig.KeyFile),
		},
		HostKeyCallback: proxy.fixedHostKeyCallback(sshConfig.KnownHost),
	}
}

func (proxy *HTTPOverSSHProxy) fixedHostKeyCallback(sshKnownHost string) ssh.HostKeyCallback {
	hostKey, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(sshKnownHost))
	return ssh.FixedHostKey(hostKey)
}

func (proxy *HTTPOverSSHProxy) publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return ssh.PublicKeys()
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return ssh.PublicKeys()
	}
	return ssh.PublicKeys(key)
}
