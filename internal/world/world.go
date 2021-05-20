package world

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"text/template"

	"github.com/jmespath/go-jmespath"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	if w.logger == nil {
		w.logger = logrus.New()
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
	azure      *Azure
	FS         FS
	Data       Data
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
	funcs := template.FuncMap(sprig.FuncMap())
	funcs["vault"] = func(path, field string) (string, error) {
		return w.Vault().Secret(path, field)
	}
	funcs["Azure"] = func(path string) (*Azure, error) {
		return w.Azure(), nil
	}
	funcs["jsonToMap"] = func(jsonData string) (map[string]interface{}, error) {
		return w.jsonToMap(jsonData)
	}
	funcs["jmsepathValue"] = func(path string, data map[string]interface{}) (interface{}, error) {
		return jmespath.Search(path, data)
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

func (w *World) jsonToMap(jsonData string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &data)
	return data, err
}
