package person_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonathanleahy/prroxy/rest-v2/internal/person"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/person/mocks"
)

func TestService_FindPerson(t *testing.T) {
	tests := []struct {
		name       string
		surname    string
		dob        string
		mockReturn *person.Person
		mockError  error
		wantErr    bool
	}{
		{
			name:    "success",
			surname: "Thompson",
			dob:     "1985-03-15",
			mockReturn: &person.Person{
				Firstname: "Emma",
				Surname:   "Thompson",
				DOB:       "1985-03-15",
				Country:   "United Kingdom",
			},
			wantErr: false,
		},
		{
			name:      "client error",
			surname:   "NotFound",
			dob:       "2000-01-01",
			mockError: errors.New("not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewPersonClient(t)
			mockClient.On("FindPerson", mock.Anything, tt.surname, tt.dob).
				Return(tt.mockReturn, tt.mockError)

			service := person.NewService(mockClient)
			ctx := context.Background()

			result, err := service.FindPerson(ctx, tt.surname, tt.dob)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.mockReturn, result)
		})
	}
}

func TestService_FindPeople(t *testing.T) {
	mockPeople := []person.Person{
		{
			Firstname: "Emma",
			Surname:   "Thompson",
			DOB:       "1985-03-15",
			Country:   "United Kingdom",
		},
		{
			Firstname: "James",
			Surname:   "Thompson",
			DOB:       "1990-05-20",
			Country:   "United States",
		},
	}

	tests := []struct {
		name       string
		surname    string
		dob        string
		mockReturn []person.Person
		mockError  error
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "success by surname",
			surname:    "Thompson",
			dob:        "",
			mockReturn: mockPeople,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "success by dob",
			surname:    "",
			dob:        "1985-03-15",
			mockReturn: mockPeople[:1],
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:      "client error",
			surname:   "Error",
			dob:       "",
			mockError: errors.New("service error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewPersonClient(t)
			mockClient.On("FindPeople", mock.Anything, tt.surname, tt.dob).
				Return(tt.mockReturn, tt.mockError)

			service := person.NewService(mockClient)
			ctx := context.Background()

			result, err := service.FindPeople(ctx, tt.surname, tt.dob)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result, tt.wantCount)
		})
	}
}
