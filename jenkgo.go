package jenkgo

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var buildExtension string = "/buildWithParameters"

type JenkinsServer struct {
	Url          *url.URL
	ApiExtension string
	User         string
	Token        string
	BaseJobPath  string
	QueryParams  map[string]interface{}
	Results      map[string]interface{}
}

func (j *JenkinsServer) callablePath() string {

	var err error

	j.Url, err = url.Parse(j.BaseJobPath)
	if err != nil {
		log.Fatal(err)
	}

	newUrl := *j.Url
	newUrl.Path += strings.TrimLeft(j.ApiExtension, "/")
	return newUrl.String()

}

func (j JenkinsServer) callApiEndpoint() {

	client := http.Client{}
	req, err := http.NewRequest("GET", j.callablePath(), nil)
	req.SetBasicAuth(j.User, j.Token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)

	if resp.StatusCode < 200 {
		log.Fatal()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal([]byte(body), &j.Results)

}

func (j *JenkinsServer) GetJob(job string) {

	j.callApiEndpoint()

	if len(j.Results) == 0 {
		log.Fatal("Didn't find any jobs")
	}

	jobList := j.Results["jobs"].([]interface{})
	var err error

	if strings.Contains(job, "/") {
		nestedJobPath := strings.Split(job, "/")
		for subdir := range nestedJobPath {
			tmpurl := matchJob(jobList, nestedJobPath[subdir])
			if tmpurl != "" {
				j.Url, err = url.Parse(tmpurl)
				j.BaseJobPath = tmpurl
				if err != nil {
					log.Fatal(err)

				}
				j.callApiEndpoint()
				jobList = j.Results["jobs"].([]interface{})
			} else {
				log.Fatal("Got invalid url")
			}
		}
	} else {
		tmpurl := matchJob(jobList, job)
		if tmpurl != "" {
			j.Url, err = url.Parse(tmpurl)
			j.BaseJobPath = tmpurl
			if err != nil {
				log.Fatal(err)

			}
		}
	}
}

func (j *JenkinsServer) GetDefaultParameters() {

	j.callApiEndpoint()

	var properties []interface{}
	for selection := range j.Results["property"].([]interface{}) {
		if val, ok := j.Results["property"].([]interface{})[selection].(map[string]interface{})["parameterDefinitions"]; ok {
			properties = val.([]interface{})
		}
	}

	for param := range properties {
		if properties[param].(map[string]interface{})["defaultParameterValue"].(map[string]interface{})["value"] != nil {
			j.QueryParams[properties[param].(map[string]interface{})["defaultParameterValue"].(map[string]interface{})["name"].(string)] = properties[param].(map[string]interface{})["defaultParameterValue"].(map[string]interface{})["value"]
		}
	}
}

func (j JenkinsServer) validateUrl() bool {

	if j.Url.Scheme != "http" && j.Url.Scheme != "https" {
		return false
	}

	_, err := url.ParseRequestURI(j.Url.String())
	return err == nil

}

func (j *JenkinsServer) TriggerJob() int {

	data := constructPath(j.QueryParams)

	j.Url.Path += buildExtension

	j.Url.RawQuery = data.Encode()

	client := http.Client{}
	req, err := http.NewRequest("POST", j.Url.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(j.User, j.Token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	return resp.StatusCode
}

func (j *JenkinsServer) GetLastBuildUrl() string {

	j.callApiEndpoint()
	return (j.Results["builds"].([]interface{})[0].(map[string]interface{})["url"].(string))

}

func NewJenkinsServer(server string, apiext string, user string, token string) *JenkinsServer {

	var err error

	j := new(JenkinsServer)
	j.Results = make(map[string]interface{})
	j.QueryParams = make(map[string]interface{})
	j.BaseJobPath = server

	j.Url, err = url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}

	j.ApiExtension = apiext

	if !j.validateUrl() {
		log.Fatal("got invalid jenkins url:", j.Url.String())
	}

	j.User = user
	j.Token = token

	return j
}

func constructPath(params map[string]interface{}) url.Values {

	val := url.Values{}

	for key, value := range params {

		switch t := value.(type) {

		case int:
			val.Set(key, strconv.Itoa(t))
		case bool:
			val.Set(key, strconv.FormatBool(t))
		case string:
			val.Set(key, t)
		}
	}

	return val

}

func matchJob(jobs []interface{}, match string) string {

	for j := range jobs {
		if match == jobs[j].(map[string]interface{})["name"] {
			return jobs[j].(map[string]interface{})["url"].(string)
		}
	}

	return ""

}

func (j *JenkinsServer) OverwriteParams(userParams map[interface{}]interface{}) {

	j.GetDefaultParameters()
	for param, value := range userParams {
		if j.QueryParams[strings.ToUpper(param.(string))] != nil {
			j.QueryParams[strings.ToUpper(param.(string))] = value
		}
	}
}
