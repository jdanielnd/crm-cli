package model

// Person represents a contact in the CRM.
type Person struct {
	ID        int64   `json:"id"`
	UUID      string  `json:"uuid"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Title     *string `json:"title"`
	Company   *string `json:"company"`
	Location  *string `json:"location"`
	Notes     *string `json:"notes"`
	Summary   *string `json:"summary"`
	OrgID     *int64  `json:"org_id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreatePersonInput holds fields for creating a person.
type CreatePersonInput struct {
	FirstName string
	LastName  *string
	Email     *string
	Phone     *string
	Title     *string
	Company   *string
	Location  *string
	Notes     *string
	OrgID     *int64
}

// UpdatePersonInput holds optional fields for updating a person.
// Pointer fields: nil = don't change, non-nil = set to this value.
type UpdatePersonInput struct {
	FirstName *string
	LastName  *string
	Email     *string
	Phone     *string
	Title     *string
	Company   *string
	Location  *string
	Notes     *string
	Summary   *string
	OrgID     *int64
}

// PersonFilters holds optional filters for listing people.
type PersonFilters struct {
	Tag   *string
	OrgID *int64
	Limit int
}

// Organization represents a company or group.
type Organization struct {
	ID        int64   `json:"id"`
	UUID      string  `json:"uuid"`
	Name      string  `json:"name"`
	Domain    *string `json:"domain"`
	Industry  *string `json:"industry"`
	Notes     *string `json:"notes"`
	Summary   *string `json:"summary"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateOrgInput holds fields for creating an organization.
type CreateOrgInput struct {
	Name     string
	Domain   *string
	Industry *string
	Notes    *string
}

// UpdateOrgInput holds optional fields for updating an organization.
type UpdateOrgInput struct {
	Name     *string
	Domain   *string
	Industry *string
	Notes    *string
	Summary  *string
}

// Interaction types
var InteractionTypes = []string{"call", "email", "meeting", "note", "message"}

// Deal stages
var DealStages = []string{"lead", "prospect", "proposal", "negotiation", "won", "lost"}

// Priority levels
var Priorities = []string{"low", "medium", "high"}

// Relationship types
var RelationshipTypes = []string{"colleague", "friend", "manager", "mentor", "referred-by"}
