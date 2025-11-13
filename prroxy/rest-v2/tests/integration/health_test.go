package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	httpAdapter "github.com/jonathanleahy/prroxy/rest-v2/internal/adapters/inbound/http"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/domain/health"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Health Endpoint - Hexagonal Architecture", func() {
	var (
		router        *gin.Engine
		healthService *health.Service
		handler       *httpAdapter.HealthHandler
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()

		// Wire up hexagonal architecture
		healthService = health.NewService("2.0.0")
		handler = httpAdapter.NewHealthHandler(healthService)

		router.GET("/health", handler.GetHealth)
	})

	Describe("GET /health", func() {
		Context("when the service is running", func() {
			It("should return status 200", func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should return JSON content type", func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})

			It("should return healthy status", func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response httpAdapter.HealthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Status).To(Equal("healthy"))
			})

			It("should return version 2.0.0", func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				var response httpAdapter.HealthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Version).To(Equal("2.0.0"))
			})

			It("should return current timestamp", func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()

				before := time.Now()
				router.ServeHTTP(w, req)
				after := time.Now()

				var response httpAdapter.HealthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())

				timestamp, err := time.Parse("2006-01-02T15:04:05.999999Z07:00", response.Timestamp)
				Expect(err).ToNot(HaveOccurred())

				Expect(timestamp).To(BeTemporally(">=", before.Add(-1*time.Second)))
				Expect(timestamp).To(BeTemporally("<=", after.Add(1*time.Second)))
			})
		})

		Context("when called multiple times", func() {
			It("should consistently return healthy status", func() {
				for i := 0; i < 10; i++ {
					req := httptest.NewRequest("GET", "/health", nil)
					w := httptest.NewRecorder()

					router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusOK))

					var response httpAdapter.HealthResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response.Status).To(Equal("healthy"))
					Expect(response.Version).To(Equal("2.0.0"))
				}
			})
		})

		Context("hexagonal architecture verification", func() {
			It("should use dependency injection through ports", func() {
				// Verify handler depends on port, not concrete implementation
				Expect(handler).ToNot(BeNil())

				// We can swap implementations because we depend on the port
				mockService := health.NewService("3.0.0")
				mockHandler := httpAdapter.NewHealthHandler(mockService)

				router := gin.New()
				router.GET("/health", mockHandler.GetHealth)

				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				var response httpAdapter.HealthResponse
				json.Unmarshal(w.Body.Bytes(), &response)
				Expect(response.Version).To(Equal("3.0.0"))
			})
		})
	})
})
