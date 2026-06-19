package model

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID"                              json:"user,omitempty"`

	Country       string     `gorm:"type:varchar(100);not null"                          json:"country"`
	OrgType       string     `gorm:"type:company_org_type;not null"                      json:"org_type"`
	Name          string     `gorm:"type:varchar(200);not null"                          json:"name"`
	INN           string     `gorm:"type:varchar(30);not null;column:inn"                json:"inn"`
	INNVerified   bool       `gorm:"not null;default:false;column:inn_verified"          json:"inn_verified"`
	INNVerifiedAt *time.Time `gorm:"column:inn_verified_at"                              json:"inn_verified_at"`

	Phone string `gorm:"type:varchar(20);not null"  json:"phone"`
	Email string `gorm:"type:varchar(100);not null" json:"email"`

	City       string  `gorm:"type:varchar(100);not null"           json:"city"`
	Region     *string `gorm:"type:varchar(100)"                    json:"region"`
	Street     string  `gorm:"type:varchar(200);not null"           json:"street"`
	PostalCode *string `gorm:"type:varchar(20);column:postal_code"  json:"postal_code"`

	RegDocURL string  `gorm:"type:text;not null;column:reg_doc_url" json:"reg_doc_url"`
	InnDocURL *string `gorm:"type:text;column:inn_doc_url"          json:"inn_doc_url"`

	Status          string     `gorm:"type:company_status;not null;default:'pending';index" json:"status"`
	RejectionReason *string    `gorm:"type:text"                                            json:"rejection_reason"`
	DocsRequestNote *string    `gorm:"type:text;column:docs_request_note"                   json:"docs_request_note"`
	ModeratorID     *uuid.UUID `gorm:"type:uuid"                                            json:"moderator_id"`
	VerifiedAt      *time.Time `json:"verified_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
