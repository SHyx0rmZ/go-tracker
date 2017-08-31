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
package tracker_test

import (
	"encoding/json"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xoebus/go-tracker"
)

var _ = Describe("Me", func() {
	It("has attributes", func() {
		var me tracker.Me
		reader := strings.NewReader(Fixture("me.json"))
		err := json.NewDecoder(reader).Decode(&me)
		Ω(err).ToNot(HaveOccurred())

		Ω(me.Username).To(Equal("vader"))
		Ω(me.Name).To(Equal("Darth Vader"))
		Ω(me.Initials).To(Equal("DV"))
		Ω(me.ID).To(Equal(101))
		Ω(me.Email).To(Equal("vader@deathstar.mil"))
	})
})

var _ = Describe("Story", func() {
	It("has attributes", func() {
		var stories []tracker.Story
		reader := strings.NewReader(Fixture("stories.json"))
		err := json.NewDecoder(reader).Decode(&stories)
		Ω(err).ToNot(HaveOccurred())
		story := stories[0]

		Ω(story.ID).Should(Equal(560))
		Ω(story.Name).Should(Equal("Tractor beam loses power intermittently"))
		Ω(story.Labels).Should(Equal([]tracker.Label{
			{ID: 10, ProjectID: 99, Name: "some-label"},
			{ID: 11, ProjectID: 99, Name: "some-other-label"},
		}))

		Ω(*story.CreatedAt).To(Equal(time.Date(2015, 07, 20, 22, 50, 50, 0, time.UTC)))
		Ω(*story.UpdatedAt).To(Equal(time.Date(2015, 07, 20, 22, 51, 50, 0, time.UTC)))
		Ω(*story.AcceptedAt).To(Equal(time.Date(2015, 07, 20, 22, 52, 50, 0, time.UTC)))
		Ω(story.Blockers).To(Equal([]tracker.Blocker{
			{ID: 12, Description: "some blocker"},
			{ID: 13, Description: "some other blocker"},
		}))
	})
})

var _ = Describe("Task", func() {
	It("has attributes", func() {
		var tasks []tracker.Task
		reader := strings.NewReader(Fixture("tasks.json"))
		err := json.NewDecoder(reader).Decode(&tasks)
		Ω(err).ToNot(HaveOccurred())
		task := tasks[0]

		Ω(task.ID).Should(Equal(52167427))
		Ω(task.StoryID).Should(Equal(137910061))
		Ω(task.Description).Should(Equal("some-task-description"))
		Ω(task.IsComplete).Should(BeTrue())
		Ω(task.Position).Should(Equal(1))
	})
})

var _ = Describe("Activity", func() {
	It("has attributes", func() {
		var activities []tracker.Activity
		reader := strings.NewReader(Fixture("activities.json"))
		err := json.NewDecoder(reader).Decode(&activities)
		Ω(err).ToNot(HaveOccurred())
		activity := activities[0]

		Ω(activity.GUID).Should(Equal("99_45"))
		Ω(activity.Message).Should(Equal("Darth Vader started this feature"))
	})
})

var _ = Describe("Project Memberships", func() {
	It("has attributes", func() {
		var projectMemberships []tracker.ProjectMembership
		reader := strings.NewReader(Fixture("project_memberships.json"))
		err := json.NewDecoder(reader).Decode(&projectMemberships)
		Ω(err).ToNot(HaveOccurred())

		membership := projectMemberships[0]
		Ω(membership.ID).Should(Equal(100))
		Ω(membership.Person.ID).Should(Equal(100))
		Ω(membership.Person.Name).Should(Equal("Emperor Palpatine"))
		Ω(membership.Person.Email).Should(Equal("emperor@galacticrepublic.gov"))
		Ω(membership.Person.Initials).Should(Equal("EP"))
		Ω(membership.Person.Username).Should(Equal("palpatine"))
	})
})
