package main

import (
	"demo/config"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type prState string

const (
	MERGED   prState = "MERGED"
	OPEN     prState = "OPEN"
	DECLINED prState = "DECLINED"
)

type PullRequest struct {
	Size       int16               `json:"size"`
	Limit      int16               `json:"limit"`
	IsLastPage bool                `json:"isLastPage"`
	Values     []PullRequestDetail `json:"values"`
}

type PullRequestDetail struct {
	Id          uint64               `json:"id"`
	Version     uint32               `json:"version"`
	Title       string               `json:"title"`
	CreatedDate uint64               `json:"createdDate"`
	UpdatedDate uint64               `json:"updatedDate"`
	ClosedDate  uint64               `json:"closedDate"`
	Author      *Author              `json:"author"`
	Reviewers   []Reviewer           `json:"reviewers"`
	State       string               `json:"state"`
	Properties  *PullRequestProperty `json:"properties"`
}

type PullRequestProperty struct {
	QgStatus          string `json:"qgStatus"`
	ResolvedTaskCount uint16 `json:"resolvedTaskCount"`
	OpenTaskCount     uint16 `json:"openTaskCount"`
	CommentCount      uint16 `json:"commentCount"`
}

type Author struct {
	U *User `json:"user"`
}
type Reviewer struct {
	Users []User `json:"users"`
}

type User struct {
	Name        string `json:"name"`
	EmailAdress string `json:"emailAdress"`
	Id          uint64 `json:"id"`
	DisplayName string `json:"displayName"`
	Active      bool   `json:"active"`
}

func getStatistics() {
	conf, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	serviceNames := strings.Split(conf.Services, ",")

	var wg sync.WaitGroup

	errChan := make(chan error, len(serviceNames))
	resultChan := make(chan PullRequest, len(serviceNames))

	for _, s := range serviceNames {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()

			// build base url
			client := &http.Client{}
			urlString := conf.BaseUrl + serviceName + "/" + "pull-requests"
			u, err := url.Parse(urlString)
			q := u.Query()
			q.Set("limit", "1000")
			q.Set("state", string(MERGED))
			u.RawQuery = q.Encode()

			//prepares request
			req, _ := http.NewRequest("GET", u.String(), strings.NewReader(u.Query().Encode()))
			req.SetBasicAuth(conf.User, conf.Pass)

			//does request
			r, err := client.Do(req)
			//in case of a problem sends it to error channel
			if err != nil {
				errChan <- err
			}
			defer r.Body.Close()

			var prBody PullRequest
			err = json.NewDecoder(r.Body).Decode(&prBody)

			//in case of a problem sends it to error channel
			if err != nil {
				errChan <- err
			}

			//take the result and send it to the result channel
			resultChan <- prBody
		}(s)
	}
	wg.Wait()
	close(errChan)
	close(resultChan)

	if len(errChan) > 0 {
		for err := range errChan {
			fmt.Println(err)
		}
	}

	for r := range resultChan {
		for _, b := range r.Values {
			if b.Author.U.Name == conf.User {
				fmt.Println("Is last page:" + strconv.FormatBool(r.IsLastPage))
				fmt.Println(b.Title)
				fmt.Println(b.Properties.CommentCount)
			}
		}
	}

}
