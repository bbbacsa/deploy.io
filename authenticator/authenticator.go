package authenticator

import (
	"crypto/md5"
	"fmt"
	"github.com/bbbacsa/deploy.io/api"
	"github.com/bbbacsa/deploy.io/vendor/code.google.com/p/gopass"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"errors"
)

type PyString string

func (py PyString) Split(str string) ( string, string , error ) {
    s := strings.Split(string(py), str)
    if len(s) < 2 {
        return "" , "", errors.New("Minimum match not found")
    }
    return s[0] , s[1] , nil
}

func Authenticate() (*api.HTTPClient, error) {
	httpClient := api.HTTPClient{GetAPIURL(), "", ""}
	err := PopulateKey(&httpClient)
	if err != nil {
		return nil, err
	}
	return &httpClient, nil
}

func PopulateKey(httpClient *api.HTTPClient) error {
	envVar := os.Getenv("DEPLOY_API_KEY")
	if envVar != "" {
		httpClient.Key = envVar
		return nil
	}

	keyFile, err := GetKeyFilePath(httpClient.BaseURL)
	if err != nil {
		return err
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		username, key, err := GetKeyByPromptingUser(*httpClient)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(keyFile, []byte(username + ":" + key), 0644); err != nil {
			return err
		}
		httpClient.Username = username
		httpClient.Key = key
	} else {
		key, err := ioutil.ReadFile(keyFile)
		if err != nil {
			return err
		}

		httpClient.Username, httpClient.Key, err = PyString(key).Split(":")
	}

	return nil
}

func GetAPIURL() string {
	apiURL := os.Getenv("DEPLOY_API_URL")

	if apiURL == "" {
		apiURL = "http://104.131.158.124:8001"
	}

	return apiURL
}

func GetKeyFilePath(baseURL string) (string, error) {
	keyDir, err := GetKeyDir()
	if err != nil {
		return "", err
	}

	h := md5.New()
	io.WriteString(h, baseURL)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	return path.Join(keyDir, hash), nil
}

func GetKeyDir() (string, error) {
	keyDir := path.Join(os.Getenv("HOME"), ".deploy", "api_keys")
	err := os.MkdirAll(keyDir, 0700)
	if err != nil {
		return "", err
	}
	return keyDir, nil
}

func GetKeyByPromptingUser(httpClient api.HTTPClient) (string, string, error) {
	username, password := Prompt()

	username, key, err := httpClient.GetAuthKey(username, password)
	if err != nil {
		return "", "", err
	}

	return username, key, nil
}

func Prompt() (string, string) {
	var (
		username string
		password string
	)
	fmt.Print("Deploy username: ")
	fmt.Scanln(&username)
	password, _ = gopass.GetPass("Password: ")
	return username, password
}
