package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Worklog struct {
	Self           string `json:"self"`
	TempoWorklogId int    `json:"tempoWorklogId"`
	JiraWorklogId  int    `json:"jiraWorklogId"`
	Issue          struct {
		Self string `json:"self"`
		Key  string `json:"key"`
		Id   int    `json:"id"`
	} `json:"issue"`
	TimeSpentSeconds int       `json:"timeSpentSeconds"`
	BillableSeconds  int       `json:"billableSeconds"`
	StartDate        string    `json:"startDate"`
	StartTime        string    `json:"startTime"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Author           struct {
		Self        string `json:"self"`
		AccountId   string `json:"accountId"`
		DisplayName string `json:"displayName"`
	} `json:"author"`
}

type WorklogCollector interface {
	GetWorklogs() []Worklog
	SetWorklogs([]Worklog)
}

type TempoRequest struct {
	Results []Worklog `json:"results"`
}

func (r TempoRequest) GetWorklogs() []Worklog {
	return r.Results
}

func (r *TempoRequest) SetWorklogs(worklog []Worklog) {
	r.Results = worklog
}

type Tempo struct {
	token string
	url   *url.URL
}

func (t *Tempo) SetToken(token string) {
	t.token = token
}

func (t *Tempo) SetUrl(urlRequest string) error {
	urlParse, err := url.Parse(urlRequest)
	if err != nil {
		return err
	}
	t.url = urlParse

	return nil
}

func (t Tempo) GetWorklogs(from time.Time, to time.Time, limit int, offset int) (WorklogCollector, error) {
	q := t.url.Query()
	q.Set("from", from.Format("2006-01-02"))
	q.Set("to", to.Format("2006-01-02"))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	t.url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, t.url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+t.token)

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	result := new(TempoRequest)

	err = json.NewDecoder(response.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t Tempo) GetAllWorklogs(from time.Time, to time.Time) (WorklogCollector, error) {
	limit := 100
	offset := 0
	var worklogsResponse WorklogCollector = &TempoRequest{}
	for {
		worklogs, err := t.GetWorklogs(from, to, limit, offset)
		if err != nil {
			return worklogsResponse, err
		}
		worklogs.GetWorklogs()
		result := worklogs.GetWorklogs()
		worklogsResponse.SetWorklogs(append(worklogsResponse.GetWorklogs(), result...))
		if len(result) != limit {
			break
		}
		offset += limit
	}

	return worklogsResponse, nil
}
