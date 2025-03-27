package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/foomo/webgrapple/pkg/httputils"
	"github.com/foomo/webgrapple/pkg/log"
	"golang.org/x/sync/errgroup"
)

type hostName string

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

func fileExistsAndIsAFile(topic string, file string) (exists bool, err error) {
	info, errStat := os.Stat(file)
	if errStat != nil && os.IsNotExist(errStat) {
		return false, nil
	}
	if info.IsDir() {
		return false, errors.New(topic + " it exists, but is a directory and not a file")
	}
	return true, nil
}

func extractDataFromURLStrings(urlStrings []string) (hosts []hostName, urls []*url.URL, err error) {
	hostMap := map[string]int{}
	urls = []*url.URL{}
	certCommonNames := map[string]struct{}{
		"localhost": {},
	}
	for _, a := range urlStrings {
		u, errParseURL := url.Parse(a)
		if errParseURL != nil {
			return nil, nil, errParseURL
		}

		hostPort := strings.Split(u.Host, ":")
		switch u.Scheme {
		case schemeHTTP:
		case schemeHTTPS:
			certCommonNames[hostPort[0]] = struct{}{}
		case "":
			return nil, nil, errors.New("empty scheme")
		default:
			return nil, nil, errors.New("unsupported scheme")
		}
		hostMap[hostPort[0]]++
		urls = append(urls, u)
	}

	hostList := []string{}
	for host := range hostMap {
		hostList = append(hostList, host)
	}
	sort.Strings(hostList)
	hosts = []hostName{}
	for _, host := range hostList {
		hosts = append(hosts, hostName(host))
	}
	return hosts, urls, nil
}

func filesExist(files ...string) (bool, error) {
	for _, f := range files {
		existsAndIsFile, err := fileExistsAndIsAFile(filepath.Base(f), f)
		if err != nil {
			return false, err
		}
		if !existsAndIsFile {
			return false, nil
		}
	}
	return true, nil
}

func ensureCertAndKey(
	l log.Logger,
	commonNames []hostName,
	certFile, keyFile string,
) (certFileCorrected, keyFileCorrected string, err error) {
	var keyExists bool
	var certExists bool

	if certFile != "" && keyFile != "" {
		certExists, err = fileExistsAndIsAFile("cert", certFile)
		if err != nil {
			return certFile, keyFile, err
		}
		keyExists, err = fileExistsAndIsAFile("key", keyFile)
		if err != nil {
			return certFile, keyFile, err
		}

		if certExists && !keyExists {
			return certFile, keyFile, errors.New("there is a cert file but no key, giving up")
		}
		if !certExists && keyExists {
			return certFile, keyFile, errors.New("there is a key file but no cert, giving up")
		}
	} else {
		certNameBase := "webgrapple-temp"
		for _, commonName := range commonNames {
			certNameBase += "-" + string(commonName)
		}
		tempDir := os.TempDir()
		certFile = filepath.Join(tempDir, "cert-"+certNameBase+".pem")
		keyFile = filepath.Join(tempDir, "key-"+certNameBase+".pem")
		l.Info(fmt.Sprintf("no key or cert given - will try temporary files at %s and %s", certFile, keyFile))
		certAndKeyExist, errFilesExist := filesExist(certFile, keyFile)
		if errFilesExist != nil {
			return certFile, keyFile, errFilesExist
		}
		certExists = certAndKeyExist
		keyExists = certAndKeyExist
	}

	if !certExists && !keyExists {
		certCommonHostNames := []string{}
		for _, certCommonName := range commonNames {
			certCommonHostNames = append(certCommonHostNames, string(certCommonName))
		}
		l.Info("generating self signed certificate and key")
		errSign := selfsign(l, certCommonHostNames, certFile, keyFile)
		if errSign != nil {
			return certFile, keyFile, errSign
		}
	} else {
		l.Info("using existing cert and key")
	}
	return certFile, keyFile, nil
}

func checkHosts(l log.Logger, hostList []hostName) map[hostName]string {
	hostAdresses := map[hostName]string{}
	for _, host := range hostList {
		addresses, errLookup := net.LookupHost(string(host))
		if errLookup != nil {
			l.Error(fmt.Sprintf("could not look up host %q, did you miss to add a hosts entry, or DNS is not available? %v", host, errLookup))
		} else {
			l.Info(fmt.Sprintf("checking host %q for addres %q", string(host), addresses))
		}
		if len(addresses) == 0 {
			l.Info("not addresses found, falling back to 127.0.0.1")
			addresses = []string{"127.0.0.1"}
		}
		hostAdresses[host] = strings.Join(addresses, ",")
	}
	return hostAdresses
}

func Run(
	ctx context.Context,
	l log.Logger,
	serviceAddress, backendURLString string,
	urlStrings []string,
	certFile, keyFile string,
	middlewareFactory WebGrappleMiddleWareCreator,
) error {
	hosts, urls, errExtractHosts := extractDataFromURLStrings(urlStrings)
	if errExtractHosts != nil {
		return errExtractHosts
	}

	hostAddresses := checkHosts(l, hosts)

	certFile, keyFile, errCertainly := ensureCertAndKey(l, hosts, certFile, keyFile)
	if errCertainly != nil {
		return errCertainly
	}

	backendURL, errParseBackendURL := url.Parse(backendURLString)
	if errParseBackendURL != nil {
		return errors.New("could not parse backend url: " + errParseBackendURL.Error())
	}

	s, errServer := newServer(backendURL, l, middlewareFactory)
	if errServer != nil {
		return errServer
	}

	usedAddressPorts := map[string]int{}
	g, gctx := errgroup.WithContext(ctx)

	for _, u := range urls {
		hostParts := strings.Split(u.Host, ":")
		hasPort := len(hostParts) > 1 && hostParts[1] != ""
		port := ""
		useTLS := false
		switch u.Scheme {
		case schemeHTTP:
			if !hasPort {
				port = ":80"
			}
		case schemeHTTPS:
			if !hasPort {
				port = ":443"
			}
			useTLS = true
		}
		addressPort := hostAddresses[hostName(hostParts[0])]
		if hasPort {
			addressPort += ":" + hostParts[1]
		} else {
			addressPort += port
		}
		usedAddressPorts[addressPort]++

		listenAddress := u.Host + port

		if usedAddressPorts[addressPort] == 1 {
			g.Go(func() error {
				name := fmt.Sprintf("proxy (%s)", u)
				httpServer := httputils.GracefulHTTPServer(gctx, l, name, listenAddress, s)
				l.Info(fmt.Sprintf("starting server on %s", addressPort))
				if useTLS {
					return httpServer.ListenAndServeTLS(certFile, keyFile)
				}
				return httpServer.ListenAndServe()
			})
		} else {
			l.Info(fmt.Sprintf("not starting server - address %q already bound", addressPort))
		}
	}

	g.Go(func() error {
		l.Info(fmt.Sprintf("starting dev client service on %q", serviceAddress))
		httpDevClient := httputils.GracefulHTTPServer(gctx, l, "dev-client", serviceAddress, s.serviceHandler)
		return httpDevClient.ListenAndServe()
	})

	return g.Wait()
}
