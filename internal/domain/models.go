package domain

import "time"

type Tenant struct {
	ID        uint64
	Name      string
	Slug      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Domain struct {
	ID           uint64
	TenantID     uint64
	Name         string
	Status       string
	DKIMSelector string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type User struct {
	ID              uint64
	TenantID        uint64
	PrimaryDomainID uint64
	LocalPart       string
	DisplayName     string
	PasswordHash    string
	Status          string
	QuotaBytes      uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
