package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	"golang.org/x/crypto/ssh"
)

// Config contains all the parameters needed to open an SSH connection
type Config struct {
	HostName  string
	KeyFile   string
	KnownHost string
	UserName  string
}

// NewConfig initializes a new ssh config structure
func NewConfig(proxyURL string, key string, knownHost string) (*Config, error) {
	if proxyURL == "" {
		return nil, nil
	}

	url, err := url.Parse("//" + proxyURL)
	if err != nil {
		return nil, err
	}
	username := "ubuntu"
	if url.User != nil {
		username = url.User.Username()
	}

	return &Config{
		HostName:  url.Host,
		KeyFile:   key,
		KnownHost: knownHost,
		UserName:  username,
	}, nil
}

// HTTPOverSSHProxy tunnels HTTP requests through SSH by opening a proxy
// and forwarding all requests
type HTTPOverSSHProxy struct{}

// Start connects through SSH to the given hostname and spins up the HTTP proxy
// which forwards all requests
func (proxy *HTTPOverSSHProxy) Start(sshConfig Config) (string, error) {
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

func (proxy *HTTPOverSSHProxy) connectToProxy(sshConfig Config) (*ssh.Client, error) {
	config := proxy.clientConfig(sshConfig)
	client, err := ssh.Dial("tcp", sshConfig.HostName, config)
	if err != nil {
		return nil, fmt.Errorf("Could not SSH into %s. Make sure that you have provided the correct SSH key and the server is a known host. Error: %s", sshConfig.HostName, err)
	}

	return client, nil
}

func (proxy *HTTPOverSSHProxy) clientConfig(sshConfig Config) *ssh.ClientConfig {
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
