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
	"fmt"
)

var DefaultURL = "https://www.pivotaltracker.com"

type Client struct {
	conn connection
}

func NewClient(token string) *Client {
	return &Client{
		conn: newConnection(token),
	}
}

func (c Client) Me() (me Me, err error) {
	request, err := c.conn.CreateRequest("GET", "/me", nil)
	if err != nil {
		return me, err
	}

	_, err = c.conn.Do(request, &me)

	return me, err
}

func (c Client) InProject(projectId int) ProjectClient {
	return ProjectClient{
		id:   projectId,
		conn: c.conn,
	}
}

func (c Client) Story(storyID int) (Story, error) {
	url := fmt.Sprintf("/stories/%d", storyID)
	request, err := c.conn.CreateRequest("GET", url, nil)
	if err != nil {
		return Story{}, err
	}

	var story Story
	_, err = c.conn.Do(request, &story)

	if err != nil {
		return Story{}, err
	}

	return story, err
}
