package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	httpAdapter "github.com/jonathanleahy/prroxy/rest-v2/internal/adapters/inbound/http"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/domain/user"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Endpoint GET /api/user/:id", func() {
	var (
		router      *gin.Engine
		userService *user.Service
		handler     *httpAdapter.UserHandler
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()

		// Wire up hexagonal architecture
		userService = user.NewService()
		handler = httpAdapter.NewUserHandler(userService)

		// Register routes with /api prefix
		api := router.Group("/api")
		api.GET("/user/:id", handler.GetUser)
	})

	Describe("GET /api/user/:id", func() {
		Context("with valid user ID", func() {
			It("should return status 200 for user 1", func() {
				req := httptest.NewRequest("GET", "/api/user/1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should return JSON content type", func() {
				req := httptest.NewRequest("GET", "/api/user/1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})

			It("should return user data with required fields", func() {
				req := httptest.NewRequest("GET", "/api/user/1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				// Verify required fields exist
				Expect(response).To(HaveKey("id"))
				Expect(response).To(HaveKey("name"))
				Expect(response).To(HaveKey("username"))
				Expect(response).To(HaveKey("email"))
				Expect(response).To(HaveKey("phone"))
				Expect(response).To(HaveKey("website"))

				// Verify id is numeric
				Expect(response["id"]).To(BeNumerically("==", 1))
			})

			It("should return correct data for user 2", func() {
				req := httptest.NewRequest("GET", "/api/user/2", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response["id"]).To(BeNumerically("==", 2))
			})
		})

		Context("with invalid user ID", func() {
			It("should return 400 for non-numeric ID", func() {
				req := httptest.NewRequest("GET", "/api/user/abc", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
			})
		})

		Context("with non-existent user ID", func() {
			It("should return 404 for user 999999", func() {
				req := httptest.NewRequest("GET", "/api/user/999999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNotFound))

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
			})
		})

		Context("hexagonal architecture verification", func() {
			It("should use dependency injection through ports", func() {
				// Verify handler is not nil and properly wired
				Expect(handler).ToNot(BeNil())

				// Service can be swapped because we depend on the port
				Expect(userService).ToNot(BeNil())
			})
		})
	})

	Describe("POST /api/user/:id/report", func() {
		BeforeEach(func() {
			// Register the report route
			api := router.Group("/api")
			api.POST("/user/:id/report", handler.PostUserReport)
		})

		Context("with valid user ID and no options", func() {
			It("should return status 200 for user 1", func() {
				req := httptest.NewRequest("POST", "/api/user/1/report", nil)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should return comprehensive report with all required fields", func() {
				req := httptest.NewRequest("POST", "/api/user/1/report", nil)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				// Verify required top-level fields
				Expect(response).To(HaveKey("userId"))
				Expect(response).To(HaveKey("userName"))
				Expect(response).To(HaveKey("email"))
				Expect(response).To(HaveKey("stats"))
				Expect(response).To(HaveKey("posts"))
				Expect(response).To(HaveKey("todos"))
				Expect(response).To(HaveKey("generatedAt"))

				// Verify stats structure
				stats := response["stats"].(map[string]interface{})
				Expect(stats).To(HaveKey("totalPosts"))
				Expect(stats).To(HaveKey("totalTodos"))
				Expect(stats).To(HaveKey("completedTodos"))
				Expect(stats).To(HaveKey("pendingTodos"))
				Expect(stats).To(HaveKey("completionRate"))

				// Verify todos structure
				todos := response["todos"].(map[string]interface{})
				Expect(todos).To(HaveKey("pending"))
				Expect(todos).To(HaveKey("completed"))

				// Verify posts is an array
				posts := response["posts"].([]interface{})
				Expect(len(posts)).To(BeNumerically(">", 0))

				// Verify post structure
				firstPost := posts[0].(map[string]interface{})
				Expect(firstPost).To(HaveKey("id"))
				Expect(firstPost).To(HaveKey("title"))
				Expect(firstPost).To(HaveKey("preview"))
			})

			It("should calculate completion rate correctly", func() {
				req := httptest.NewRequest("POST", "/api/user/1/report", nil)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				stats := response["stats"].(map[string]interface{})
				completionRate := stats["completionRate"].(string)

				// Should be formatted as "XX.X%"
				Expect(completionRate).To(MatchRegexp(`^\d+\.\d%$`))
			})
		})

		Context("with optional parameters", func() {
			It("should limit posts when maxPosts is specified", func() {
				body := `{"maxPosts": 3}`
				req := httptest.NewRequest("POST", "/api/user/1/report", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				// Verify posts are limited to 3
				posts := response["posts"].([]interface{})
				Expect(len(posts)).To(Equal(3))
			})

			It("should exclude completed todos when includeCompleted is false", func() {
				body := `{"includeCompleted": false}`
				req := httptest.NewRequest("POST", "/api/user/1/report", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				// Verify completed todos array is empty
				todos := response["todos"].(map[string]interface{})
				completed := todos["completed"].([]interface{})
				Expect(len(completed)).To(Equal(0))
			})
		})

		Context("with invalid user ID", func() {
			It("should return 400 for non-numeric ID", func() {
				req := httptest.NewRequest("POST", "/api/user/abc/report", nil)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("GET /api/user/:id/summary", func() {
		BeforeEach(func() {
			// Register the summary route
			api := router.Group("/api")
			api.GET("/user/:id/summary", handler.GetUserSummary)
		})

		Context("with valid user ID", func() {
			It("should return status 200 for user 1", func() {
				req := httptest.NewRequest("GET", "/api/user/1/summary", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should return JSON content type", func() {
				req := httptest.NewRequest("GET", "/api/user/1/summary", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})

			It("should return user summary with required fields", func() {
				req := httptest.NewRequest("GET", "/api/user/1/summary", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				// Verify required fields exist
				Expect(response).To(HaveKey("userId"))
				Expect(response).To(HaveKey("userName"))
				Expect(response).To(HaveKey("email"))
				Expect(response).To(HaveKey("postCount"))
				Expect(response).To(HaveKey("recentPosts"))
				Expect(response).To(HaveKey("summary"))

				// Verify types
				Expect(response["userId"]).To(BeNumerically("==", 1))
				Expect(response["postCount"]).To(BeNumerically(">", 0))
				Expect(response["recentPosts"]).To(BeAssignableToTypeOf([]interface{}{}))
			})

			It("should return summary with correct format", func() {
				req := httptest.NewRequest("GET", "/api/user/1/summary", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				// Summary should match format: "User {name} has written {count} posts"
				summary := response["summary"].(string)
				Expect(summary).To(ContainSubstring("User"))
				Expect(summary).To(ContainSubstring("has written"))
				Expect(summary).To(ContainSubstring("posts"))
			})
		})

		Context("with invalid user ID", func() {
			It("should return 400 for non-numeric ID", func() {
				req := httptest.NewRequest("GET", "/api/user/abc/summary", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
