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

import "time"

type Me Person

type Person struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
	ID       int    `json:"id"`
	Email    string `json:"email"`
}

type Project struct {
	Id int
}

type Story struct {
	ID        int `json:"id,omitempty"`
	ProjectID int `json:"project_id,omitempty"`

	URL string `json:"url,omitempty"`

	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Type        StoryType  `json:"story_type,omitempty"`
	State       StoryState `json:"current_state,omitempty"`

	Labels []Label `json:"labels,omitempty"`

	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

type Task struct {
	ID      int `json:"id,omitempty"`
	StoryID int `json:"story_id,omitempty"`

	Description string `json:"description,omitempty"`
	IsComplete  bool   `json:"complete,omitempty"`
	Position    int    `json:"position,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type Comment struct {
	Text string `json:"text,omitempty"`
}

type Blocker struct {
	ID          int    `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}

type Label struct {
	ID        int `json:"id,omitempty"`
	ProjectID int `json:"project_id,omitempty"`

	Name string `json:"name"`
}

type StoryType string

const (
	StoryTypeFeature = "feature"
	StoryTypeBug     = "bug"
	StoryTypeChore   = "chore"
	StoryTypeRelease = "release"
)

type StoryState string

const (
	StoryStateUnscheduled = "unscheduled"
	StoryStatePlanned     = "planned"
	StoryStateStarted     = "started"
	StoryStateFinished    = "finished"
	StoryStateDelivered   = "delivered"
	StoryStateAccepted    = "accepted"
	StoryStateRejected    = "rejected"
)

type Activity struct {
	Kind             string        `json:"kind"`
	GUID             string        `json:"guid"`
	ProjectVersion   int           `json:"project_version"`
	Message          string        `json:"message"`
	Highlight        string        `json:"highlight"`
	Changes          []interface{} `json:"changes"`
	PrimaryResources []interface{} `json:"primary_resources"`
	Project          interface{}   `json:"project"`
	PerformedBy      interface{}   `json:"performed_by"`
	OccurredAt       time.Time     `json:"occurred_at"`
}

type ProjectMembership struct {
	ID     int `json:"id"`
	Person Person
}
