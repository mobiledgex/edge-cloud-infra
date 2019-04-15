package mexos

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func GetURIFile(uri string) ([]byte, error) {
	log.DebugLog(log.DebugLevelMexos, "attempt to get uri file", "uri", uri)
	// if _, err := url.ParseRequestURI(uri); err != nil {
	// 	return nil, err
	// }
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		res, err := GetHTTPFile(uri)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "error getting http uri file", "uri", uri, "error", err)
			return nil, err
		}
		return res, nil
	}
	if strings.HasPrefix(uri, "scp://") {
		res, err := GetSCPFile(uri)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "error getting scp uri file", "uri", uri, "error", err)
			return nil, err
		}
		return res, nil
	}
	if strings.HasPrefix(uri, "file:///") {
		uri = strings.Replace(uri, "file:///", "", -1)
	}
	// if err := validateDomain(uri); err != nil {
	// 	return ioutil.ReadFile(uri)
	// }
	log.DebugLog(log.DebugLevelMexos, "attempt to read uri as normal file", "uri", uri)
	res, err := ioutil.ReadFile(uri)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "error getting file uri file", "uri", uri, "error", err)
		return nil, err
	}
	return res, nil
}

func GetHTTPFile(uri string) ([]byte, error) {
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

func GetSCPFile(uri string) ([]byte, error) {
	log.DebugLog(log.DebugLevelMexos, "attempt to get scp uri file", "uri", uri)
	part1 := strings.Replace(uri, "scp://", "mobiledgex@", -1)
	slashindex := strings.Index(part1, "/")
	if slashindex < 0 {
		return nil, fmt.Errorf("malformed uri, missing /")
	}
	addr := part1[:slashindex]
	if len(part1) < (slashindex + 1) {
		return nil, fmt.Errorf("malformed uri, too short")
	}
	fn := part1[slashindex+1:]
	if len(fn) < 1 {
		return nil, fmt.Errorf("malformed uri, fn too short")
	}
	return sh.Command("ssh", "-o", sshOpts[0], "-o", sshOpts[1], "-i", PrivateSSHKey(), addr, "cat", fn).Output()
}

// func CopyURIFile(uri string, fn string) error {
// 	res, err := GetURIFile(mf, uri)
// 	if err != nil {
// 		return err
// 	}
// 	err = ioutil.WriteFile(fn, res, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func GetFileNameWithExt(fileUrlPath string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get file name with extension from url", "file-url", fileUrlPath)
	fileUrl, err := url.Parse(fileUrlPath)
	if err != nil {
		return "", fmt.Errorf("Error parsing file URL %s, %v", fileUrlPath, err)
	}

	path := fileUrl.Path
	segments := strings.Split(path, "/")

	return segments[len(segments)-1], nil
}

func GetFileName(fileUrlPath string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get file name from url", "file-url", fileUrlPath)
	fileName, err := GetFileNameWithExt(fileUrlPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(fileName, filepath.Ext(fileName)), nil
}

func GetArtifactoryCreds() (string, string, error) {
	af_user := os.Getenv("MEX_ARTIFACTORY_USER")
	if af_user == "" {
		return "", "", fmt.Errorf("Env variable MEX_ARTIFACTORY_USER not set")
	}
	af_pass := os.Getenv("MEX_ARTIFACTORY_PASS")
	if af_pass == "" {
		return "", "", fmt.Errorf("Env variable MEX_ARTIFACTORY_PASS not set")
	}
	return af_user, af_pass, nil
}

func GetUrlUpdatedTime(fileUrlPath string) (time.Time, error) {
	log.DebugLog(log.DebugLevelMexos, "get url last-modified time", "file-url", fileUrlPath)
	af_user, af_pass, err := GetArtifactoryCreds()
	if err != nil {
		return time.Time{}, err
	}
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", fileUrlPath, nil)
	req.SetBasicAuth(af_user, af_pass)
	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("Error fetching last modified time of URL %s, %v", fileUrlPath, err)
	}
	tStr := resp.Header.Get("Last-modified")
	return time.Parse(time.RFC1123, tStr)
}

func DownloadFile(fileUrlPath string) error {
	log.DebugLog(log.DebugLevelMexos, "attempt to download file", "file-url", fileUrlPath)
	fileUrl, err := url.Parse(fileUrlPath)
	if fileUrl.Host == "artifactory.mobiledgex.net" {
		af_user, af_pass, err := GetArtifactoryCreds()
		if err != nil {
			return err
		}
		_, err = sh.Command("wget", "--user", af_user, "--password", af_pass, fileUrlPath, sh.Dir("/tmp")).Output()
	} else {
		_, err = sh.Command("wget", "--no-check-certificate", fileUrlPath, sh.Dir("/tmp")).Output()
	}
	return err
}

func DeleteFile(filePath string) error {
	var err error
	if _, err = os.Stat(filePath); !os.IsNotExist(err) {
		_, err = sh.Command("rm", filePath).Output()
	}
	return err
}
