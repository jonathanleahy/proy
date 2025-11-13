package person

// Person represents a person entity with basic information
type Person struct {
	Firstname string `json:"firstname"`
	Surname   string `json:"surname"`
	DOB       string `json:"dob"`     // Date of birth in YYYY-MM-DD format
	Country   string `json:"country"` // Must be "country" (singular) to match v1 API
}
