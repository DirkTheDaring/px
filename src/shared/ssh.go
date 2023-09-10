package shared

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetResource(url string) (string, error) {

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "http get error (%s): %v\n", url, err)
		return "", err
	}
	defer resp.Body.Close()

	str, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func RemoveCR(input string) string {
	return strings.ReplaceAll(input, "\r", "")

}
func ConvertStringToArray(input string) []string {

	return strings.Split(input, "\n")
}

func GetSSHResource(uri string) []string {
	//u, _ := url.Parse(uri)
	//fmt.Fprintf(os.Stderr, "url:%v\nscheme:%v host:%v Path:%v\n", u, u.Scheme, u.Host, u.Path)
	str, err := GetResource(uri)
	if err != nil {
		return []string{}
	}
	str = RemoveCR(str)
	array := ConvertStringToArray(str)

	newArray := []string{}

	for _, item := range array {
		if !strings.HasPrefix(item, "ssh-") {
			continue
		}
		newArray = append(newArray, item)
	}
	return newArray
}
