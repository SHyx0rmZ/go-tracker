package tracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type IterationScope string

const (
	IterationScopeDone IterationScope = "done"
	IterationScopeCurrent IterationScope = "current"
	IterationScopeBacklog IterationScope = "backlog"
	IterationScopeCurrentBacklog IterationScope = "current_backlog"
	IterationScopeDoneCurrent IterationScope = "done_current"
)

type ProjectClient struct {
	id   int
	conn connection
}

type Iteration struct {
	Number int `json:"number"`
	ProjectID int `json:"project_id"`
	Length int `json:"length"`
	TeamStrength float32 `json:"team_strength"`
	Stories []Story `json:"stories"`
	Start time.Time `json:"start"`
	Finish time.Time `json:"finish"`
	Velocity float32 `json:"velocity"`
	Points int `json:"points"`
	AcceptedPoints int `json:"accepted_points"`
	EffectivePoints float32 `json:"effective_points"`
	Accepted *json.RawMessage `json:"accepted"`
	Created *json.RawMessage `json:"created"`
	Analytics *json.RawMessage `json:"analytics"`
	Kind string `json:"kind"`
}

type IterationsQuery struct {
	Scope IterationScope
	Label  string

	Limit  int
	Offset int
}

func (query IterationsQuery) Query() url.Values {
	params := url.Values{}

	if query.Scope != "" {
		params.Set("scope", string(query.Scope))
	}

	if query.Label != "" {
		params.Set("label", query.Label)
	}

	if query.Limit != 0 {
		params.Set("limit", fmt.Sprintf("%d", query.Limit))
	}

	if query.Offset != 0 {
		params.Set("offset", fmt.Sprintf("%d", query.Offset))
	}

	return params
}


func (p ProjectClient) Iterations(query IterationsQuery) ([]Iteration, Pagination, error) {
	request, err := p.createRequest("GET", "/iterations", query.Query())
	if err != nil {
		return nil, Pagination{}, err
	}

	var iterations []Iteration
	pagination, err := p.conn.Do(request, &iterations)
	if err != nil {
		return nil, Pagination{}, err
	}

	return iterations, pagination, err
}

func (p ProjectClient) Stories(query StoriesQuery) ([]Story, Pagination, error) {
	request, err := p.createRequest("GET", "/stories", query.Query())
	if err != nil {
		return nil, Pagination{}, err
	}

	var stories []Story
	pagination, err := p.conn.Do(request, &stories)
	if err != nil {
		return nil, Pagination{}, err
	}

	return stories, pagination, err
}

func (p ProjectClient) StoryActivity(storyId int, query ActivityQuery) (activities []Activity, err error) {
	url := fmt.Sprintf("/stories/%d/activity", storyId)

	request, err := p.createRequest("GET", url, query.Query())
	if err != nil {
		return activities, err
	}

	_, err = p.conn.Do(request, &activities)
	return activities, err
}

func (p ProjectClient) StoryTasks(storyId int, query TaskQuery) (tasks []Task, err error) {
	url := fmt.Sprintf("/stories/%d/tasks", storyId)

	request, err := p.createRequest("GET", url, query.Query())
	if err != nil {
		return tasks, err
	}

	_, err = p.conn.Do(request, &tasks)
	return tasks, err
}

func (p ProjectClient) StoryComments(storyId int, query CommentsQuery) (comments []Comment, err error) {
	url := fmt.Sprintf("/stories/%d/comments", storyId)

	request, err := p.createRequest("GET", url, query.Query())
	if err != nil {
		return comments, err
	}

	_, err = p.conn.Do(request, &comments)
	return comments, err
}

func (p ProjectClient) DeliverStoryWithComment(storyId int, comment string) error {
	err := p.DeliverStory(storyId)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/stories/%d/comments", storyId)
	request, err := p.createRequest("POST", url, nil)
	if err != nil {
		return err
	}

	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(Comment{
		Text: comment,
	})

	p.addJSONBodyReader(request, buffer)

	_, err = p.conn.Do(request, nil)
	return err
}

func (p ProjectClient) DeliverStory(storyId int) error {
	url := fmt.Sprintf("/stories/%d", storyId)
	request, err := p.createRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	p.addJSONBody(request, `{"current_state":"delivered"}`)

	_, err = p.conn.Do(request, nil)
	return err
}

func (p ProjectClient) CreateStory(story Story) (Story, error) {
	request, err := p.createRequest("POST", "/stories", nil)
	if err != nil {
		return Story{}, err
	}

	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(story)

	p.addJSONBodyReader(request, buffer)

	var createdStory Story
	_, err = p.conn.Do(request, &createdStory)
	return createdStory, err
}

func (p ProjectClient) UpdateStory(story Story) (Story, error) {
	url := fmt.Sprintf("/stories/%d", story.ID)
	request, err := p.createRequest("PUT", url, nil)
	if err != nil {
		return Story{}, err
	}

	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(story)

	p.addJSONBodyReader(request, buffer)

	var updatedStory Story
	_, err = p.conn.Do(request, &updatedStory)
	return updatedStory, nil
}

func (p ProjectClient) DeleteStory(storyId int) error {
	url := fmt.Sprintf("/stories/%d", storyId)
	request, err := p.createRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	_, err = p.conn.Do(request, nil)
	return err
}

func (p ProjectClient) CreateTask(storyID int, task Task) (Task, error) {
	url := fmt.Sprintf("/stories/%d/tasks", storyID)
	request, err := p.createRequest("POST", url, nil)
	if err != nil {
		return Task{}, err
	}

	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(task)

	p.addJSONBodyReader(request, buffer)

	var createdTask Task
	_, err = p.conn.Do(request, &createdTask)
	return createdTask, err
}

func (p ProjectClient) CreateComment(storyID int, comment Comment) (Comment, error) {
	url := fmt.Sprintf("/stories/%d/comments", storyID)
	request, err := p.createRequest("POST", url, nil)
	if err != nil {
		return Comment{}, err
	}

	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(comment)

	p.addJSONBodyReader(request, buffer)

	var createdComment Comment
	_, err = p.conn.Do(request, &createdComment)
	return createdComment, err
}

func (p ProjectClient) CreateBlocker(storyID int, blocker Blocker) (Blocker, error) {
	url := fmt.Sprintf("/stories/%d/blockers", storyID)
	request, err := p.createRequest("POST", url, nil)
	if err != nil {
		return Blocker{}, err
	}

	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(blocker)

	p.addJSONBodyReader(request, buffer)

	var createdBlocker Blocker
	_, err = p.conn.Do(request, &createdBlocker)
	return createdBlocker, err
}

func (p ProjectClient) ProjectMemberships() ([]ProjectMembership, error) {
	request, err := p.createRequest("GET", "/memberships", nil)
	if err != nil {
		return []ProjectMembership{}, err
	}

	var memberships []ProjectMembership
	_, err = p.conn.Do(request, &memberships)
	if err != nil {
		return []ProjectMembership{}, err
	}

	return memberships, nil
}

func (p ProjectClient) createRequest(method string, path string, params url.Values) (*http.Request, error) {
	projectPath := fmt.Sprintf("/projects/%d%s", p.id, path)
	return p.conn.CreateRequest(method, projectPath, params)
}

func (p ProjectClient) addJSONBodyReader(request *http.Request, body io.Reader) {
	request.Header.Add("Content-Type", "application/json")
	request.Body = ioutil.NopCloser(body)
}

func (p ProjectClient) addJSONBody(request *http.Request, body string) {
	p.addJSONBodyReader(request, strings.NewReader(body))
}
