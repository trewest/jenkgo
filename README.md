# Jenkgo


## About
---
Jenkgo is a simple packge used to trigger parameterized Jenkins jobs. Features of the client include:

- Authenticating with Jenkins REST API using basic auth
- Finding nested and un-nested jobs by name
- Executing with default or custom parameters

## Examples
### Executing with default parameters
```go
func main() {

	jenkinsUrl := "http://127.0.0.1:8080/"
	jenkinApiExt := "/api/json"
	user := "trewest"
	token := "<user-api-token>"
	jobname := "jenkgo-job"

	jenkins := jenkgo.NewJenkinsServer(jenkinsUrl, jenkinApiExt, user, token)

	jenkins.GetJob(jobname)
    jenkins.GetDefaultParameters()
	jenkins.TriggerJob()

}
```
### Job output

![default output](../assets/jenkgo-default-output.png)

### Executing with custom parameters
```go
func main() {

	jenkinsUrl := "http://127.0.0.1:8080/"
	jenkinApiExt := "/api/json"
	user := "trewest"
	token := "<user-api-token>"
	jobname := "jenkgo-job"
	customParameters := map[interface{}]interface{}{
		"MESSAGE":         "Hello, Jenkgo!",
		"FAVORITE_NUMBER": "25",
	}

	jenkins := jenkgo.NewJenkinsServer(jenkinsUrl, jenkinApiExt, user, token)

	jenkins.GetJob(jobname)
	jenkins.OverwriteParams(customParameters)
	jenkins.TriggerJob()

}
```
### Job output

![custom output](../assets/jenkgo-custom-output.png)
