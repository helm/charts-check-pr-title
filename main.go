/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	gin "gopkg.in/gin-gonic/gin.v1"
	//"golang.org/x/oauth2"
)

var (
	// The shared secret between the app and GitHub
	// This application only works with one shared secret. It was designed for
	// a single repo (and single repo situation). The app should be modified
	// for a multi-repo context where each repo can have its own secret
	sharedSecret string

	// The name of the repo e.g. "foo/bar"
	repoFullName string

	// The token for the user/bot that will be updating the label and sending
	// a notification
	ghToken string

	// A regext looking for patterns like [stable/mariadb] and [test]
	//TODO(mattfarina): Should we check for docs in the title?
	retestRe = regexp.MustCompile(`^(\[.*\/.*]|\[test\]).*$`)

	// The body of the comment to post
	cmtBody = `Thank you for submitting the pull request. There are many people who review pull requests for the different charts and tests. To help us review your pull request would you consider updating the pull request title to the format:

 * **[<repo>/<chart>] title** (e.g., _[stable/mariadb] title_) if this pull request is for a specific chart
 * **[test] title** if this pull request is for the common tests
 
`
)

func main() {

	// Get config from environment
	sharedSecret = os.Getenv("GITHUB_SHARED_SECRET")
	repoFullName = os.Getenv("GITHUB_REPO_NAME")
	ghToken = os.Getenv("GITHUB_TOKEN")

	// Disable color in output
	gin.DisableConsoleColor()

	router := gin.New()

	// Recovery enables Gin to handle panics and provides a 500 error
	router.Use(gin.Recovery())

	// gin.Default() setups up recovery and logging on all paths. In this case
	// we want to skip /healthz checks so as not to clutter up the logs.
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/healthz"))

	// We can use this to check the health or and make sure the app is online
	router.GET("/healthz", healthz)

	router.POST("/webhook", processHook)

	router.Run()
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

func processHook(c *gin.Context) {

	// Validate payload
	sig := c.GetHeader("X-Hub-Signature")
	if sig == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing X-Hub-Signature"})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logit("ERROR: Failed to read request body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	if err := validateSig(sig, body); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Validating payload against signature failed"})
		return
	}

	// Check the event type. Make sure just a PR
	// We need to get the event from the Request object as Gin in the middle
	// does some normalization that breaks this particular header name.
	event := c.Request.Header.Get("X-GitHub-Event")

	// We are only interested in pull requests
	if event != "pull_request" {
		c.JSON(http.StatusOK, gin.H{"message": "Skipping event type"})
		return
	}

	// Get the payload body as an object
	e, err := github.ParseWebHook(event, body)
	if err != nil {
		logit("ERROR: Failed to parse body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Malformed body"})
		return
	}

	payload := e.(*github.PullRequestEvent)

	// Filter by repo name
	if *payload.Repo.FullName != repoFullName {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Payload submitted for wrong repo"})
		return
	}

	// Filter pull request actions we aren't intersted in like labels being added/removed
	if *payload.Action != "opened" {
		c.JSON(http.StatusOK, gin.H{"message": "Skipping action"})
		return
	}

	// Check the title of the PR and add comment if label missing proper format
	if !validTitle(*payload.PullRequest.Title) {

		// Leave a comment
		ctx, client := ghClient()
		cmt := &github.PullRequestComment{
			Body: &cmtBody,
		}
		parts := strings.Split(*payload.Repo.FullName, "/")
		_, resp, err := client.PullRequests.CreateComment(ctx, parts[0], parts[1], *payload.Number, cmt)
		if err != nil {
			logit("ERROR: Failed to post comment to %d: %s", *payload.Number, err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to post comment"})
			return
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			logit("ERROR: Failed to post comment to %d. Status code: &s", *payload.Number, resp.Status)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to post comment"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func validateSig(sig string, body []byte) error {
	digest := hmac.New(sha1.New, []byte(sharedSecret))
	digest.Write(body)
	sum := digest.Sum(nil)
	checksum := fmt.Sprintf("sha1=%x", sum)
	if subtle.ConstantTimeCompare([]byte(checksum), []byte(sig)) != 1 {
		logit("ERROR: Expected signature %q, but got %q", checksum, sig)
		return errors.New("payload signature check failed")
	}
	return nil
}

func logit(message string, vars ...interface{}) {
	fmt.Fprintf(gin.DefaultWriter, "[APP] "+message+"\n", vars...)
}

func ghClient() (context.Context, *github.Client) {
	t := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})
	c := context.Background()
	tc := oauth2.NewClient(c, t)
	return c, github.NewClient(tc)
}

func validTitle(title string) bool {

	// Check the structure
	found := retestRe.FindAllStringSubmatch(title, -1)
	fmt.Printf("%q\n", found)
	if len(found) == 0 || len(found[0]) < 2 {
		return false
	}

	// TODO(mattfarina): check if the chart name actually exists? This could be done with a GET request

	return true
}
