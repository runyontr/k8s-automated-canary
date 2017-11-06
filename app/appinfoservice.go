package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type AppInfoService interface {
	GetAppInfo() (AppInfo, error)
}


func New() (AppInfoService, error) {

	return &appInfoBaseline{}, nil

}


//appInfoBaseline is the implementation of the AppInfoService interface.  This implementation has a bug where
// the Namespace value is not populated.
type appInfoBaseline struct {
}

//GetAppInfo returns the app info of the running application
func (s *appInfoBaseline) GetAppInfo() (AppInfo, error) {

	info := AppInfo{}
	info.Labels = make(map[string]string)

	info.PodName = os.Getenv("MY_POD_NAME") //custom defined in the deployment spec
	//time.Sleep(3*time.Second)
	//info.Namespace = os.Getenv("MY_POD_NAMESPACE") //custom defined in the deployment spec

	file, err := os.Open("/etc/labels")
	if err != nil {
		return info, err
	}
	defer file.Close()

	//overkill, but read it fresh each time
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')

		// check if the line has = sign
		// and process the line. Ignore the rest.
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}

				value = strings.Replace(value, "\"", "", -1)
				switch key {
				case "app":
					info.AppName = value
				case "release":
					info.Release = value
				default:
					info.Labels[key] = value
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return info, err
		}
	}

	return info, err
}

