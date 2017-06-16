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
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"

	"github.com/xoebus/go-tracker"
)

var _ = Describe("Tracker Client", func() {
	var (
		server *ghttp.Server
		client *tracker.Client
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		tracker.DefaultURL = server.URL()
		client = tracker.NewClient("api-token")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("getting information about the current user", func() {
		var statusCode int

		It("works if everything goes to plan", func() {
			statusCode = http.StatusOK

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/services/v5/me"),
				verifyTrackerToken(),

				ghttp.RespondWith(statusCode, Fixture("me.json")),
			))

			me, err := client.Me()

			Ω(err).ToNot(HaveOccurred())
			Ω(me.Username).To(Equal("vader"))
		})

		It("returns an error if the response is not successful", func() {
			statusCode = http.StatusInternalServerError

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(statusCode, ""),
			))

			client := tracker.NewClient("api-token")
			_, err := client.Me()
			Ω(err).To(MatchError("request failed (500)"))
		})

		It("returns a helpful error if the token is invalid", func() {
			statusCode = http.StatusUnauthorized

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(statusCode, ""),
			))

			client := tracker.NewClient("api-token")
			_, err := client.Me()
			Ω(err).To(MatchError("invalid token"))
		})

		It("returns an error if the request fails", func() {
			server.Close()

			client := tracker.NewClient("api-token")
			_, err := client.Me()

			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(MatchRegexp("failed to make request"))
			server = ghttp.NewServer()
		})

		It("returns an error if the request can't be created", func() {
			tracker.DefaultURL = "aaaaa)#Q&%*(*"

			client := tracker.NewClient("api-token")
			_, err := client.Me()

			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(MatchRegexp("failed to create request"))
		})

		It("returns an error if the response JSON is broken", func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, `{"`),
			))

			client := tracker.NewClient("api-token")
			_, err := client.Me()

			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(MatchRegexp("invalid json response"))
		})
	})

	Describe("listing stories", func() {
		It("gets all the stories by default", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/services/v5/projects/99/stories"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("stories.json")),
				),
			)

			client := tracker.NewClient("api-token")

			stories, pagination, err := client.InProject(99).Stories(tracker.StoriesQuery{})
			Ω(stories).Should(HaveLen(4))
			Ω(pagination).Should(BeZero())
			Ω(err).ToNot(HaveOccurred())
		})

		It("returns pagination info allowing the caller to follow through pages themselves", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/services/v5/projects/99/stories"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("stories.json"), http.Header{
						"X-Tracker-Pagination-Total":    []string{"1"},
						"X-Tracker-Pagination-Offset":   []string{"2"},
						"X-Tracker-Pagination-Limit":    []string{"3"},
						"X-Tracker-Pagination-Returned": []string{"4"},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/services/v5/projects/99/stories", "offset=1234"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("stories.json"), http.Header{
						"X-Tracker-Pagination-Total":    []string{"5"},
						"X-Tracker-Pagination-Offset":   []string{"6"},
						"X-Tracker-Pagination-Limit":    []string{"7"},
						"X-Tracker-Pagination-Returned": []string{"8"},
					}),
				),
			)

			client := tracker.NewClient("api-token")

			stories, pagination, err := client.InProject(99).Stories(tracker.StoriesQuery{})
			Ω(stories).Should(HaveLen(4))
			Ω(pagination).Should(Equal(tracker.Pagination{
				Total:    1,
				Offset:   2,
				Limit:    3,
				Returned: 4,
			}))
			Ω(err).ToNot(HaveOccurred())

			stories, pagination, err = client.InProject(99).Stories(tracker.StoriesQuery{Offset: 1234})
			Ω(stories).Should(HaveLen(4))
			Ω(pagination).Should(Equal(tracker.Pagination{
				Total:    5,
				Offset:   6,
				Limit:    7,
				Returned: 8,
			}))
			Ω(err).ToNot(HaveOccurred())
		})

		It("allows different queries to be made", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/services/v5/projects/99/stories", "with_state=finished"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("stories.json")),
				),
			)

			client := tracker.NewClient("api-token")

			query := tracker.StoriesQuery{
				State: tracker.StoryStateFinished,
			}
			stories, pagination, err := client.InProject(99).Stories(query)
			Ω(stories).Should(HaveLen(4))
			Ω(pagination).Should(BeZero())
			Ω(err).ToNot(HaveOccurred())
		})
	})

	Describe("listing project memberships", func() {
		It("gets all the project memberships", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/services/v5/projects/99/memberships"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("project_memberships.json")),
				),
			)

			client := tracker.NewClient("api-token")

			memberships, err := client.InProject(99).ProjectMemberships()
			Ω(memberships).Should(HaveLen(7))
			Ω(err).ToNot(HaveOccurred())
		})
	})

	Describe("listing a story's activity", func() {
		It("gets the story's activity", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/services/v5/projects/99/stories/560/activity"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("activities.json")),
				),
			)

			client := tracker.NewClient("api-token")

			activities, err := client.InProject(99).StoryActivity(560, tracker.ActivityQuery{})
			Ω(activities).Should(HaveLen(4))
			Ω(err).ToNot(HaveOccurred())
		})

		It("allows different queries to be made", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						"/services/v5/projects/99/stories/560/activity",
						"limit=2&occurred_after=1000000000000&occurred_before=1433091819000&offset=1&since_version=1",
					),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("activities.json")),
				),
			)

			client := tracker.NewClient("api-token")

			query := tracker.ActivityQuery{
				Limit:          2,
				Offset:         1,
				OccurredBefore: 1433091819000,
				OccurredAfter:  1000000000000,
				SinceVersion:   1,
			}
			activities, err := client.InProject(99).StoryActivity(560, query)
			Ω(activities).Should(HaveLen(4))
			Ω(err).ToNot(HaveOccurred())
		})
	})

	Describe("listing a story's tasks", func() {
		It("gets the story's tasks", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/services/v5/projects/99/stories/560/tasks"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("tasks.json")),
				),
			)

			client := tracker.NewClient("api-token")

			tasks, err := client.InProject(99).StoryTasks(560, tracker.TaskQuery{})
			Ω(tasks).Should(HaveLen(3))
			Ω(err).ToNot(HaveOccurred())
		})

		It("allows different queries to be made", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						"/services/v5/projects/99/stories/560/activity",
						"limit=2&occurred_after=1000000000000&occurred_before=1433091819000&offset=1&since_version=1",
					),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, Fixture("activities.json")),
				),
			)

			client := tracker.NewClient("api-token")

			query := tracker.ActivityQuery{
				Limit:          2,
				Offset:         1,
				OccurredBefore: 1433091819000,
				OccurredAfter:  1000000000000,
				SinceVersion:   1,
			}
			activities, err := client.InProject(99).StoryActivity(560, query)
			Ω(activities).Should(HaveLen(4))
			Ω(err).ToNot(HaveOccurred())
		})
	})

	Describe("delivering a story", func() {
		It("HTTP PUTs it in its place", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/services/v5/projects/99/stories/15225523"),
					ghttp.VerifyJSON(`{"current_state":"delivered"}`),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, ""),
				),
			)

			client := tracker.NewClient("api-token")

			err := client.InProject(99).DeliverStory(15225523)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("HTTP PUTs it in its place with a comment", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/services/v5/projects/99/stories/15225523"),
					ghttp.VerifyJSON(`{"current_state":"delivered"}`),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/services/v5/projects/99/stories/15225523/comments"),
					ghttp.VerifyJSON(`{"text":"some delive\"}ry comment with tricky text"}`),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusCreated, ""),
				),
			)

			client := tracker.NewClient("api-token")

			comment := `some delive"}ry comment with tricky text`
			err := client.InProject(99).DeliverStoryWithComment(15225523, comment)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("creating a story", func() {
		It("POSTs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/services/v5/projects/99/stories"),
					ghttp.VerifyJSON(`{"name":"Exhaust ports are ray shielded"}`),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, `{
						"id": 1234,
						"project_id": 5678,
						"name": "Exhaust ports are ray shielded",
						"url": "https://some-url.biz/1234"
					}`),
				),
			)

			client := tracker.NewClient("api-token")

			story, err := client.InProject(99).CreateStory(tracker.Story{
				Name: "Exhaust ports are ray shielded",
			})
			Ω(story).Should(Equal(tracker.Story{
				ID:        1234,
				ProjectID: 5678,

				Name: "Exhaust ports are ray shielded",

				URL: "https://some-url.biz/1234",
			}))
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("deleting a story", func() {
		It("DELETES", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/services/v5/projects/99/stories/1234"),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, ""),
				),
			)
			client := tracker.NewClient("api-token")
			err := client.InProject(99).DeleteStory(1234)
			Ω(err).ShouldNot(HaveOccurred())
		})
		Context("when the delete is not successful", func() {
			It("returns error saying request failed", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/services/v5/projects/99/stories/1234"),
						verifyTrackerToken(),

						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)
				client := tracker.NewClient("api-token")
				err := client.InProject(99).DeleteStory(1234)
				Ω(err).Should(Equal(errors.New("request failed (500)")))
			})
		})
	})

	Describe("creating a task", func() {
		It("POSTs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/services/v5/projects/99/stories/560/tasks"),
					ghttp.VerifyJSON(`{"description":"some-tracker-task", "position":1}`),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, `{
					  "kind": "task",
					  "id": 1234,
					  "story_id": 560,
					  "description": "some-tracker-task",
					  "complete": false,
					  "position": 1
					}`),
				),
			)

			client := tracker.NewClient("api-token")

			task, err := client.InProject(99).CreateTask(560, tracker.Task{
				Description: "some-tracker-task",
				Position:    1,
			})

			Ω(task).Should(Equal(tracker.Task{
				ID:          1234,
				StoryID:     560,
				Description: "some-tracker-task",
				Position:    1,
				IsComplete:  false,
			}))
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("creating a blocker", func() {
		It("POSTs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/services/v5/projects/99/stories/560/blockers"),
					ghttp.VerifyJSON(`{"description":"some-tracker-blocker"}`),
					verifyTrackerToken(),

					ghttp.RespondWith(http.StatusOK, `{
			          "kind": "blocker",
                      "id": 111,
                      "story_id": 560,
                      "description": "some-tracker-blocker",
                      "person_id": 101,
                      "created_at": "2017-03-07T12:00:00Z",
                      "updated_at": "2017-03-07T12:00:00Z"
					}`),
				),
			)

			client := tracker.NewClient("api-token")

			blocker, err := client.InProject(99).CreateBlocker(560, tracker.Blocker{
				Description: "some-tracker-blocker",
			})

			Expect(blocker).Should(Equal(tracker.Blocker{
				ID: 111,
				Description: "some-tracker-blocker",
			}))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})

func verifyTrackerToken() http.HandlerFunc {
	headers := http.Header{
		"X-TrackerToken": {"api-token"},
	}

	return ghttp.VerifyHeader(headers)
}
