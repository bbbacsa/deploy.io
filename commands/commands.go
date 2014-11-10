package commands

import (
	"errors"
	"flag"
	"fmt"
	"github.com/bbbacsa/deploy.io/api"
	"github.com/bbbacsa/deploy.io/authenticator"
//	"github.com/bbbacsa/deploy.io/proxy"
//	"github.com/bbbacsa/deploy.io/tlsconfig"
	"github.com/bbbacsa/deploy.io/utils"
//	"github.com/bbbacsa/deploy.io/vendor/crypto/tls"
	"io/ioutil"
//	"net"
	"os"
	"os/exec"
//	"os/signal"
	"path"
	"strings"
//	"syscall"
	"text/tabwriter"
)

type Command struct {
	Run       func(cmd *Command, args []string) error
	UsageLine string
	Short     string
	Long      string
	Flag      flag.FlagSet
}

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

func (c *Command) UsageError(format string, args ...interface{}) error {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\nUsage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
	return fmt.Errorf(format, args...)
}

var All = []*Command{
	Docker,
	Hosts,
	IP,
	Proxy,
	Run,
}

var HostSubcommands = []*Command{
	CreateHost,
	RemoveHost,
}

func init() {
	Hosts.Run = RunHosts
	CreateHost.Run = RunCreateHost
	RemoveHost.Run = RunRemoveHost
	Docker.Run = RunDocker
	IP.Run = RunIP
}

var Hosts = &Command{
	UsageLine: "hosts",
	Short:     "Manage hosts",
	Long: `Manage hosts.

Usage: deploy hosts [COMMAND] [ARGS...]

Commands:
  ls          List hosts (default)
  create      Create a host
  rm          Remove a host

Run 'deploy hosts COMMAND -h' for more information on a command.
`,
}

var CreateHost = &Command{
	UsageLine: "create [-m MEMORY] [NAME]",
	Short:     "Create a host",
	Long: fmt.Sprintf(`Create a host.

You can optionally specify a name for the host - if not, it will be
named 'default', and 'deploy docker' commands will use it automatically.

You can also specify how much RAM the host should have with -m.
Valid amounts are %s.`, validSizes),
}

var flCreateSize = CreateHost.Flag.String("m", "512M", "")
var validSizes = "512M, 1G, 2G, 4G and 8G"

var RemoveHost = &Command{
	UsageLine: "rm [-f] [NAME]",
	Short:     "Remove a host",
	Long: `Remove a host.

You can optionally specify which host to remove - if you don't, the default
host (named 'default') will be removed.

Set -f to bypass the confirmation step, at your peril.
`,
}

var flRemoveHostForce = RemoveHost.Flag.Bool("f", false, "")

var Docker = &Command{
	UsageLine: "docker [-H HOST] [COMMAND...]",
	Short:     "Run a Docker command against a host",
	Long: `Run a Docker command against a host.

Wraps the 'docker' command-line tool - see the Docker website for reference:

    http://docs.docker.io/en/latest/reference/commandline/

You can optionally specify a host by name - if you don't, the default host
will be used.`,
}

var flDockerHost = Docker.Flag.String("H", "", "")

var Proxy = &Command{
	UsageLine: "proxy [-H HOST] [LISTEN_URL]",
	Short:     "Start a local proxy to a host's Docker daemon",
	Long: `Start a local proxy to a host's Docker daemon.

By default, listens on a Unix socket at a random path, e.g.

    $ deploy proxy
    Started proxy at unix:///tmp/deploy-12345/deploy.sock

    $ docker -H unix:///tmp/deploy-12345/deploy.sock run ubuntu echo hello world
    hello world

Instead, you can specify a URL to listen on, which can be a socket or TCP address:

    $ deploy proxy unix:///path/to/socket
    $ deploy proxy tcp://localhost:1234
`,
}

var flProxyHost = Proxy.Flag.String("H", "", "")

var IP = &Command{
	UsageLine: "ip [NAME]",
	Short:     "Print a hosts's IP address to stdout",
	Long: `Print a hosts's IP address to stdout.

You can optionally specify which host - if you don't, the default
host (named 'default') will be assumed.
`,
}

var Run = &Command{
	UsageLine: "run [-H HOST] COMMAND [ARGS...]",
	Short:     "Run a command with the DOCKER_HOST envvar set",
	Long: `Start a proxy to a Deploy.IO host and run a command locally
with the DOCKER_HOST environment variable set.

For example:

$ deploy run fig up

You can optionally specify which host - if you don't, the default
host (named 'default') will be assumed.
`,
}

var flRunHost = Run.Flag.String("H", "", "")

func RunHosts(cmd *Command, args []string) error {
	list := len(args) == 0 || (len(args) == 1 && args[0] == "ls")

	if !list {
		for _, subcommand := range HostSubcommands {
			if subcommand.Name() == args[0] {
				subcommand.Flag.Usage = func() { subcommand.Usage() }
				subcommand.Flag.Parse(args[1:])
				args = subcommand.Flag.Args()
				err := subcommand.Run(subcommand, args)
				return err
			}
		}

		return fmt.Errorf("Unknown `hosts` subcommand: %s", args[0])
	}

	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	hosts, err := httpClient.GetHosts()
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSIZE\tIP")
	for _, host := range hosts {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", host.ID, host.Name, utils.HumanSize(host.Size*1024*1024), host.IPAddress)
	}
	writer.Flush()

	return nil
}

func RunCreateHost(cmd *Command, args []string) error {
	if len(args) > 1 {
		return cmd.UsageError("`deploy hosts create` expects at most 1 argument, but got more: %s", strings.Join(args[1:], " "))
	}

	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	hostName, humanName := GetHostName(args)
	humanName = utils.Capitalize(humanName)

	size, sizeString := GetHostSize()
	if size == -1 {
		fmt.Fprintf(os.Stderr, "Sorry, %q isn't a size we support.\nValid sizes are %s.\n", sizeString, validSizes)
		return nil
	}

	host, err := httpClient.CreateHost(hostName, size)
	if err != nil {
		// HACK. api.go should decode JSON and return a specific type of error for this case.
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintf(os.Stderr, "%s is already running.\nYou can create additional hosts with `deploy hosts create [NAME]`.\n", humanName)
			return nil
		}
		if strings.Contains(err.Error(), "Invalid value") {
			fmt.Fprintf(os.Stderr, "Sorry, '%s' isn't a valid host name.\nHost names can only contain lowercase letters, numbers and underscores.\n", hostName)
			return nil
		}
		if strings.Contains(err.Error(), "Unsupported size") {
			fmt.Fprintf(os.Stderr, "Sorry, %q isn't a size we support.\nValid sizes are %s.\n", sizeString, validSizes)
			return nil
		}

		return err
	}
	fmt.Fprintf(os.Stderr, "%s running at %s\n", humanName, host.IPAddress)

	return nil
}

func RunRemoveHost(cmd *Command, args []string) error {
	if len(args) > 1 {
		return cmd.UsageError("`deploy hosts rm` expects at most 1 argument, but got more: %s", strings.Join(args[1:], " "))
	}

	hostName, humanName := GetHostName(args)

	if !*flRemoveHostForce {
		var confirm string
		fmt.Printf("Going to remove %s. All data on it will be lost.\n", humanName)
		fmt.Print("Are you sure you're ready? [yN] ")
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) != "y" {
			return nil
		}
	}

	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	err = httpClient.DeleteHost(hostName)
	if err != nil {
		// HACK. api.go should decode JSON and return a specific type of error for this case.
		if strings.Contains(err.Error(), "Not found") {
			fmt.Fprintf(os.Stderr, "%s doesn't seem to be running.\nYou can view your running hosts with `deploy hosts`.\n", utils.Capitalize(humanName))
			return nil
		}

		return err
	}
	fmt.Fprintf(os.Stderr, "Removed %s\n", humanName)

	return nil
}

func RunDocker(cmd *Command, args []string) error {
	return WithDockerProxy("", *flDockerHost, func(listenURL string, ca string, clientCert string, clientKey string) error {
	  args = append([]string{"--tlscacert='" + ca + "'"}, args...)
	  args = append([]string{"--tlscert='" + clientCert + "'"}, args...)
	  args = append([]string{"--tlskey='" + clientKey + "'"}, args...)
		err := CallDocker(args, listenURL)
		if err != nil {
			return fmt.Errorf("Docker exited with error")
		}
		return nil
	})
}

func RunIP(cmd *Command, args []string) error {
	if len(args) > 1 {
		return cmd.UsageError("`deploy ip` expects at most 1 argument, but got more: %s", strings.Join(args[1:], " "))
	}

	hostName, _ := GetHostName(args)

	host, err := GetHost(hostName)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, host.IPAddress)
	return nil
}

func WithDockerProxy(listenURL, hostName string, callback func(string, string, string, string) error) error {
	if hostName == "" {
		hostName = "default"
	}

	host, err := GetHost(hostName)
	if err != nil {
		return err
	}

	listenURL = "tcp://" + host.IPAddress + ":" + fmt.Sprintf("%d", host.Port)
	
	ca, caerr := ioutil.TempFile(os.TempDir(), "ca")
	cert, certerr := ioutil.TempFile(os.TempDir(), "cert")
	key, keyerr := ioutil.TempFile(os.TempDir(), "key")
	
	if caerr != nil {
	  return caerr
	}
	if certerr != nil {
	  return certerr
	}
	if keyerr != nil {
  	return keyerr
	}

	caerr = ioutil.WriteFile(ca.Name(), []byte(deployCerts), 0x644)
	certerr = ioutil.WriteFile(cert.Name(), []byte(host.ClientCert), 0x644)
	keyerr = ioutil.WriteFile(key.Name(), []byte(host.ClientKey), 0x644)
	
	if err := callback(listenURL, ca.Name(), cert.Name(), key.Name()); err != nil {
	  os.Remove(ca.Name())
	  os.Remove(cert.Name())
	  os.Remove(key.Name())
		return err
	}
	

	return nil
}

func CallDocker(args []string, dockerHost string) error {
	dockerPath := GetDockerPath()
	if dockerPath == "" {
		return errors.New("Can't find `docker` executable in $PATH.\nYou might need to install it: http://docs.docker.io/en/latest/installation/#installation-list")
	}

	os.Setenv("DOCKER_HOST", dockerHost)
	os.Setenv("DOCKER_TLS_VERIFY", "1")

	cmd := exec.Command(dockerPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetDockerPath() string {
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		dockerPath := path.Join(dir, "docker")
		_, err := os.Stat(dockerPath)
		if err == nil {
			return dockerPath
		}
	}
	return ""
}

func GetHostName(args []string) (string, string) {
	hostName := "default"

	if len(args) > 0 {
		hostName = args[0]
	}

	return hostName, GetHumanHostName(hostName)
}

func GetHumanHostName(hostName string) string {
	if hostName == "default" {
		return "default host"
	} else {
		return fmt.Sprintf("host '%s'", hostName)
	}
}

func GetHostSize() (int, string) {
	sizeString := *flCreateSize

	bytes, err := utils.RAMInBytes(sizeString)
	if err != nil {
		return -1, sizeString
	}

	megs := bytes / (1024 * 1024)
	if megs < 1 {
		return -1, sizeString
	}

	return int(megs), sizeString
}

func GetHost(hostName string) (*api.Host, error) {
	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return nil, err
	}

	host, err := httpClient.GetHost(hostName)
	if err != nil {
		// HACK. api.go should decode JSON and return a specific type of error for this case.
		if strings.Contains(err.Error(), "Not found") {
			humanName := GetHumanHostName(hostName)
			return nil, fmt.Errorf("%s doesn't seem to be running.\nYou can create it with `deploy hosts create %s`.", utils.Capitalize(humanName), hostName)
		}

		return nil, err
	}

	return host, nil
}

var deployCerts string = `-----BEGIN CERTIFICATE-----
MIIEKTCCAxGgAwIBAgIJAP81C5xoXHunMA0GCSqGSIb3DQEBCwUAMIGqMQswCQYD
VQQGEwJQSDEPMA0GA1UECAwGTGFndW5hMRAwDgYDVQQHDAdDYWxhbWJhMS4wLAYD
VQQKDCVCcnljaGVUZWNoIEludGVybmV0IFNvbHV0aW9ucyBDb21wYW55MQwwCgYD
VQQLDANEZXYxGDAWBgNVBAMMDzEwNC4xMzEuMTU4LjEyNDEgMB4GCSqGSIb3DQEJ
ARYRYmJiYWNzYUBnbWFpbC5jb20wHhcNMTQxMTEwMDUyMzU5WhcNMTUxMTEwMDUy
MzU5WjCBqjELMAkGA1UEBhMCUEgxDzANBgNVBAgMBkxhZ3VuYTEQMA4GA1UEBwwH
Q2FsYW1iYTEuMCwGA1UECgwlQnJ5Y2hlVGVjaCBJbnRlcm5ldCBTb2x1dGlvbnMg
Q29tcGFueTEMMAoGA1UECwwDRGV2MRgwFgYDVQQDDA8xMDQuMTMxLjE1OC4xMjQx
IDAeBgkqhkiG9w0BCQEWEWJiYmFjc2FAZ21haWwuY29tMIIBIjANBgkqhkiG9w0B
AQEFAAOCAQ8AMIIBCgKCAQEAzWBb0yQS5ca0dQWhsrPdFcsfUGazhJ8EXM+2Np5s
bj7wiT06TSunB+ME1Aj61KKxb9gI1QSW8LJy9Xp/1R1r7SWJ5VAAb8oSXP92w0Dk
ph8MXPl1x8K3B22hJk0jiIADdS09AG30cp7osW6uqz7ARgsQh4khh2DohaB0zM1t
9uLrDgUdP9BAVlFVRSYpfKMBPZ5PfmKmYod9GYOA9/Nxs44N/PhmvFMI42cVoL88
YFZ3x/U7Iu495Hri9fJ1roHAh7Z7nGL3sD/iGd3bGXOeDztXWiqp59qSFUyvOuyC
K/paZnz5izBlaz6Zir+M+zMcjAh4qvccfWyRlNZWFtAONwIDAQABo1AwTjAdBgNV
HQ4EFgQUSBE1/wLJOO3R2TFs8katsJy9B+0wHwYDVR0jBBgwFoAUSBE1/wLJOO3R
2TFs8katsJy9B+0wDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEATZbL
PdPIH8ahQpGdtTb0shcxOYcuLcrn67kxEzeOkXcOsmHhw8RdWkQiglMPBwdyi0Xj
8yY9ri1Jv1jC/swAAmtsB6qd+oxJaiVn2G+okVX2xXaLCQROwfIcNEnwVxUXyNwG
hMZEiNT1kymy8MI5FwQqZ4hvbbUqcMSrB2O1z5C8zwDL2eXm8LjrmRkRpb+pP9fX
kwYPbQO4v0v4PKge2ezhWc4u0WFN3Zg68XS2YB5anKQzK1heFiB79mbHyRKF+t9c
F4Un6peNMm7WBxup68KTBCQb6lK6jhtTUvVirMCjwaXUQOHOrRT9QzoBrfgJm0OJ
gCAxbCdK3lcDQKxC8Q==
-----END CERTIFICATE-----`