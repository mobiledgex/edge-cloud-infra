package mexos

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"
)

//validateDomain does strange validation, not strictly domain, due to the data passed from controller.
// if it is FQDN it is valid. And if it starts with http:// or https:// and followed by fqdn, it is valid.
func validateDomain(uri string) error {
	if isDomainName(uri) {
		return nil
	}
	fqdn := uri2fqdn(uri)
	if isDomainName(fqdn) {
		return nil
	}
	return fmt.Errorf("URI %s is not a valid domain name", uri)
}

func GetURIFile(mf *Manifest, uri string) ([]byte, error) {
	log.DebugLog(log.DebugLevelMexos, "attempt to get uri file", "uri", uri)
	if err := validateDomain(uri); err != nil {
		return ioutil.ReadFile(uri)
	}
	if _, err := url.ParseRequestURI(uri); err != nil {
		return nil, err
	}
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return GetHTTPFile(mf, uri)
	}
	if strings.HasPrefix(uri, "scp://") {
		return GetSCPFile(mf, uri)
	}
	return nil, fmt.Errorf("unsupported uri %s", uri)
}

func GetHTTPFile(mf *Manifest, uri string) ([]byte, error) {
	log.DebugLog(log.DebugLevelMexos, "attempt to get http uri file", "uri", uri)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, fmt.Errorf("http status not OK, %v", resp.StatusCode)
}

func GetSCPFile(mf *Manifest, uri string) ([]byte, error) {
	log.DebugLog(log.DebugLevelMexos, "attempt to get scp uri file", "uri", uri)
	addr := strings.Replace(uri, "scp://", "mobiledgex@", -1)
	return sh.Command("ssh", "-o", sshOpts[0], "-o", sshOpts[1], "-i", PrivateSSHKey(), addr, "cat").Output()
}

func CopyURIFile(mf *Manifest, uri string, fn string) error {
	res, err := GetURIFile(mf, uri)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fn, res, 0644)
	if err != nil {
		return err
	}
	return nil
}
