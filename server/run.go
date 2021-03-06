package server

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
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
	logger *zap.Logger,
	commonNames []hostName,
	certFile, keyFile string,
) (certFileCorrected, keyFileCorrected string, err error) {
	certExists, keyExists := false, false

	if certFile != "" && keyFile != "" {
		certExists, errCertExists := fileExistsAndIsAFile("cert", certFile)
		if errCertExists != nil {
			return certFile, keyFile, errCertExists
		}
		keyExists, errKeyExists := fileExistsAndIsAFile("key", keyFile)
		if errKeyExists != nil {
			return certFile, keyFile, errKeyExists
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
		logger.Info("no key or cert given - will try temporary files", zap.String("cert", certFile), zap.String("key", keyFile))
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
		logger.Info("generating self signed certificate and key", zap.Strings("host-names", certCommonHostNames))
		errSign := selfsign(logger, certCommonHostNames, certFile, keyFile)
		if errSign != nil {
			return certFile, keyFile, errSign
		}
	} else {
		logger.Info("using existing cert and key")
	}
	return certFile, keyFile, nil

}

func checkHosts(logger *zap.Logger, hostList []hostName) (hostAdresses map[hostName]string) {
	hostAdresses = map[hostName]string{}
	for _, host := range hostList {
		addresses, errLookup := net.LookupHost(string(host))
		if errLookup != nil {
			logger.Error("could not look up host, did you miss to add a hosts entry, or DNS is not available?", zap.String("host", string(host)), zap.Error(errLookup))
		} else {
			logger.Info("host check", zap.String("host", string(host)), zap.Strings("addresses", addresses))
		}
		if len(addresses) == 0 {
			logger.Info("not addresses found, falling back to 127.0.0.1")
			addresses = []string{"127.0.0.1"}
		}
		hostAdresses[host] = strings.Join(addresses, ",")
	}
	return hostAdresses
}

func run(logger *zap.Logger, serviceAddress, backendURLString string, urlStrings []string, certFile, keyFile string) error {

	hosts, urls, errExtractHosts := extractDataFromURLStrings(urlStrings)
	if errExtractHosts != nil {
		return errExtractHosts
	}

	hostAddresses := checkHosts(logger, hosts)

	certFile, keyFile, errCertainly := ensureCertAndKey(logger, hosts, certFile, keyFile)
	if errCertainly != nil {
		return errCertainly
	}

	backendURL, errParseBackendURL := url.Parse(backendURLString)
	if errParseBackendURL != nil {
		return errors.New("could not parse backend url: " + errParseBackendURL.Error())
	}

	s, errServer := newServer(backendURL, logger)
	if errServer != nil {
		return errServer
	}

	chanErr := make(chan error)

	usedAddressPorts := map[string]int{}

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

		zapAddressport := zap.String("address:port", addressPort)
		if usedAddressPorts[addressPort] == 1 {
			go func(u *url.URL, useTLS bool, port string) {
				logger.Info("starting server", zapAddressport, zap.Bool("useTLS", useTLS))
				if useTLS {
					chanErr <- http.ListenAndServeTLS(u.Host+port, certFile, keyFile, s)
				} else {
					chanErr <- http.ListenAndServe(u.Host+port, s)
				}
			}(u, useTLS, port)
		} else {
			logger.Info("not starting server - address already bound", zapAddressport)
		}
	}
	go func() {
		logger.Info("starting dev client service", zap.String("address", serviceAddress))
		chanErr <- http.ListenAndServe(serviceAddress, s.serviceHandler)
	}()
	return <-chanErr
}
