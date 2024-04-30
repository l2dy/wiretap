package udp

import (
	"errors"
	"net"
	"net/url"
	"os"
	"sync"

	"github.com/l2dy/socks5"
	"golang.org/x/net/proxy"
)

var (
	allProxyEnv = &envOnce{
		names: []string{"ALL_PROXY", "all_proxy"},
	}

	errNoProxy = errors.New("no proxy")
)

// dialProxyFromEnvironment dials a proxy server from the environment.
// If no proxy environment variables are set, it returns a nil error and net.Conn.
func dialProxyFromEnvironment(raddr *net.UDPAddr) (net.Conn, error) {
	allProxy := allProxyEnv.Get()
	if len(allProxy) == 0 {
		return nil, nil
	}

	u, err := url.Parse(allProxy)
	if err != nil {
		return nil, nil
	}

	var auth proxy.Auth
	if u.User != nil {
		auth.User = u.User.Username()
		if p, ok := u.User.Password(); ok {
			auth.Password = p
		}
	}

	switch u.Scheme {
	case "socks5", "socks5h":
		addr := u.Hostname()
		port := u.Port()
		if port == "" {
			port = "1080"
		}
		// UDP timeout should be set by caller
		c, err := socks5.NewClient(net.JoinHostPort(addr, port), auth.User, auth.Password, 60, 0)
		if err != nil {
			return nil, err
		}
		return c.Dial("udp", raddr.String())
	}

	return nil, nil
}

// envOnce looks up an environment variable (optionally by multiple
// names) once. It mitigates expensive lookups on some platforms
// (e.g. Windows).
// (Borrowed from net/http/transport.go)
type envOnce struct {
	names []string
	once  sync.Once
	val   string
}

func (e *envOnce) Get() string {
	e.once.Do(e.init)
	return e.val
}

func (e *envOnce) init() {
	for _, n := range e.names {
		e.val = os.Getenv(n)
		if e.val != "" {
			return
		}
	}
}

// reset is used by tests
func (e *envOnce) reset() {
	e.once = sync.Once{}
	e.val = ""
}
