// Copyright 2016 Christopher Brown. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
)

type ProjectClient struct {
	id   int
	conn connection
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
	return updatedStory, err
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
