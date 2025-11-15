// Package person provides person lookup functionality.
// It handles person data retrieval from the external user service.
package person

// Person represents a person from rest-external-user service
type Person struct {
	Firstname string `json:"firstname"`
	Surname   string `json:"surname"`
	DOB       string `json:"dob"` // Date of birth in format YYYY-MM-DD
	Country   string `json:"country"`
}
