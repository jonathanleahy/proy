package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Person struct {
	Firstname string `json:"firstname"`
	Surname   string `json:"surname"`
	DOB       string `json:"dob"`
	Country   string `json:"country"`
}

// In-memory data store
var people = []Person{
	{Firstname: "Emma", Surname: "Thompson", DOB: "1985-03-15", Country: "United Kingdom"},
	{Firstname: "James", Surname: "Anderson", DOB: "1990-07-22", Country: "United States"},
	{Firstname: "Sophie", Surname: "Martinez", DOB: "1988-11-08", Country: "Spain"},
	{Firstname: "Liam", Surname: "O'Connor", DOB: "1992-01-30", Country: "Ireland"},
	{Firstname: "Isabella", Surname: "Rossi", DOB: "1987-05-14", Country: "Italy"},
	{Firstname: "Noah", Surname: "Schmidt", DOB: "1991-09-03", Country: "Germany"},
	{Firstname: "Olivia", Surname: "Dubois", DOB: "1989-12-20", Country: "France"},
	{Firstname: "William", Surname: "Johnson", DOB: "1986-04-17", Country: "Canada"},
	{Firstname: "Ava", Surname: "Kowalski", DOB: "1993-08-25", Country: "Poland"},
	{Firstname: "Lucas", Surname: "Silva", DOB: "1984-02-11", Country: "Brazil"},
	{Firstname: "Mia", Surname: "Andersson", DOB: "1995-06-09", Country: "Sweden"},
	{Firstname: "Alexander", Surname: "Petrov", DOB: "1988-10-27", Country: "Russia"},
	{Firstname: "Charlotte", Surname: "Van Der Berg", DOB: "1990-03-05", Country: "Netherlands"},
	{Firstname: "Benjamin", Surname: "Kim", DOB: "1987-07-18", Country: "South Korea"},
	{Firstname: "Amelia", Surname: "Nguyen", DOB: "1992-11-12", Country: "Vietnam"},
	{Firstname: "Ethan", Surname: "Wilson", DOB: "1989-01-24", Country: "Australia"},
	{Firstname: "Harper", Surname: "Brown", DOB: "1991-05-31", Country: "New Zealand"},
	{Firstname: "Mason", Surname: "Garcia", DOB: "1986-09-14", Country: "Mexico"},
	{Firstname: "Evelyn", Surname: "Hansen", DOB: "1994-12-07", Country: "Denmark"},
	{Firstname: "Logan", Surname: "Murphy", DOB: "1988-04-21", Country: "Ireland"},
	{Firstname: "Aria", Surname: "Patel", DOB: "1990-08-16", Country: "India"},
	{Firstname: "Sebastian", Surname: "Müller", DOB: "1985-02-28", Country: "Austria"},
	{Firstname: "Luna", Surname: "Sato", DOB: "1993-06-03", Country: "Japan"},
	{Firstname: "Jack", Surname: "Taylor", DOB: "1987-10-19", Country: "United Kingdom"},
	{Firstname: "Chloe", Surname: "Leblanc", DOB: "1991-12-26", Country: "Canada"},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3006"
	}

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "rest-external-user",
			"port":    port,
		})
	})

	// Person lookup endpoint
	router.GET("/person", func(c *gin.Context) {
		surname := c.Query("surname")
		dob := c.Query("dob")

		// If neither parameter provided, return error
		if surname == "" && dob == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Missing required parameters",
				"message": "At least one of surname or dob is required",
			})
			return
		}

		// Search for matches
		var matches []Person
		for _, person := range people {
			match := true
			if surname != "" && person.Surname != surname {
				match = false
			}
			if dob != "" && person.DOB != dob {
				match = false
			}
			if match {
				matches = append(matches, person)
			}
		}

		// If both surname and dob provided, return single result or 404
		if surname != "" && dob != "" {
			if len(matches) > 0 {
				c.JSON(http.StatusOK, matches[0])
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Person not found",
					"message": "No person found with surname \"" + surname + "\" and dob \"" + dob + "\"",
				})
			}
			return
		}

		// Return array for partial searches
		c.JSON(http.StatusOK, matches)
	})

	log.Printf("Starting rest-external-user service on port %s...", port)
	log.Printf("Endpoints:")
	log.Printf("  GET /health")
	log.Printf("  GET /person?surname=X&dob=YYYY-MM-DD")
	log.Printf("Serving %d people", len(people))

	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
