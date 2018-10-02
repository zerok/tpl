package world

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var ErrInsecureRequired = errors.New("This feature requires the --insecure flag")

type Options struct {
	Logger     *logrus.Logger
	Insecure   bool
	LeftDelim  string
	RightDelim string
}

// New generates ... a new world ...
func New(opts *Options) *World {
	if opts == nil {
		opts = &Options{}
	}
	w := &World{
		logger:     opts.Logger,
		leftDelim:  opts.LeftDelim,
		rightDelim: opts.RightDelim,
		insecure:   opts.Insecure,
	}
	return w
}

// Env lazily loads environment variables.
func (w *World) Env() Env {
	if w.env == nil {
		env := Env{}
		for _, kv := range os.Environ() {
			elems := strings.SplitN(kv, "=", 2)
			env[elems[0]] = elems[1]
		}
		w.env = &env
	}
	return *w.env
}

// World acts as a container for all the knowledge we want to expose through
// the template.
type World struct {
	logger     *logrus.Logger
	Network    Network
	env        *Env
	vault      *Vault
	FS         FS
	leftDelim  string
	rightDelim string
	insecure   bool
}

// Render takes a template stream as input and converts the world's knowledge
// through that template into output written to the output stream.
func (w *World) Render(out io.Writer, in io.Reader) error {
	rawTmpl, err := ioutil.ReadAll(in)
	if err != nil {
		return errors.Wrap(err, "failed to read template")
	}
	tmpl, err := template.New("ROOT").Delims(w.leftDelim, w.rightDelim).Funcs(w.Funcs()).Parse(string(rawTmpl))
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}
	return tmpl.Execute(out, w)
}

func (w *World) Funcs() template.FuncMap {
	funcs := template.FuncMap{}
	funcs["vault"] = func(path, field string) string {
		return w.Vault().Secret(path, field)
	}

	for name, fn := range sprig.FuncMap() {
		funcs[name] = fn
	}

	return funcs
}

// Network contains knowledge about the local network.
type Network struct {
	externalIP string
}

// ExternalIP attempts to determine the host's IP address used to connect
// to external hosts.
func (nw *Network) ExternalIP() string {
	if nw.externalIP != "" {
		return nw.externalIP
	}
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return fmt.Sprintf("<ERR: %s>", err.Error())
	}
	defer conn.Close()
	addr := conn.LocalAddr()
	var ip string
	switch a := addr.(type) {
	case *net.IPAddr:
		ip = a.IP.String()
	case *net.UDPAddr:
		ip = a.IP.String()
	case *net.TCPAddr:
		ip = a.IP.String()
	default:
		return fmt.Sprintf("<ERR: Unsupported format of %v>", addr)
	}
	nw.externalIP = ip
	return ip
}
