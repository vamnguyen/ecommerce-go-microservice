package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserProfile struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FirstName string
	LastName  string
	Phone     string
	Address   string
	City      string
	Country   string
	PostalCode string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUserProfile(userID uuid.UUID, firstName, lastName, phone, address, city, country, postalCode string) *UserProfile {
	now := time.Now()
	return &UserProfile{
		ID:         uuid.New(),
		UserID:     userID,
		FirstName:  firstName,
		LastName:   lastName,
		Phone:      phone,
		Address:    address,
		City:       city,
		Country:    country,
		PostalCode: postalCode,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (p *UserProfile) Update(firstName, lastName, phone, address, city, country, postalCode string) {
	p.FirstName = firstName
	p.LastName = lastName
	p.Phone = phone
	p.Address = address
	p.City = city
	p.Country = country
	p.PostalCode = postalCode
	p.UpdatedAt = time.Now()
}

