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

// Interaction represents an activity log entry.
type Interaction struct {
	ID         int64   `json:"id"`
	UUID       string  `json:"uuid"`
	Type       string  `json:"type"`
	Subject    *string `json:"subject"`
	Content    *string `json:"content"`
	Direction  *string `json:"direction"`
	OccurredAt string  `json:"occurred_at"`
	PersonIDs  []int64 `json:"person_ids"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// CreateInteractionInput holds fields for creating an interaction.
type CreateInteractionInput struct {
	Type       string
	Subject    *string
	Content    *string
	Direction  *string
	OccurredAt *string
	PersonIDs  []int64
}

// InteractionFilters holds optional filters for listing interactions.
type InteractionFilters struct {
	PersonID *int64
	Type     *string
	Limit    int
}

// Deal represents a sales opportunity.
type Deal struct {
	ID        int64    `json:"id"`
	UUID      string   `json:"uuid"`
	Title     string   `json:"title"`
	Value     *float64 `json:"value"`
	Stage     string   `json:"stage"`
	PersonID  *int64   `json:"person_id"`
	OrgID     *int64   `json:"org_id"`
	Notes     *string  `json:"notes"`
	ClosedAt  *string  `json:"closed_at"`
	Archived  bool     `json:"-"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// CreateDealInput holds fields for creating a deal.
type CreateDealInput struct {
	Title    string
	Value    *float64
	Stage    string
	PersonID *int64
	OrgID    *int64
	Notes    *string
}

// UpdateDealInput holds optional fields for updating a deal.
type UpdateDealInput struct {
	Title    *string
	Value    *float64
	Stage    *string
	PersonID *int64
	OrgID    *int64
	Notes    *string
	ClosedAt *string
}

// DealFilters holds optional filters for listing deals.
type DealFilters struct {
	Stage         *string
	PersonID      *int64
	OrgID         *int64
	ExcludeClosed bool
	Limit         int
}

// PipelineStage represents a summary of deals in a stage.
type PipelineStage struct {
	Stage      string  `json:"stage"`
	Count      int     `json:"count"`
	TotalValue float64 `json:"total_value"`
}

// Tag represents a label that can be applied to any entity.
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Task represents a follow-up or to-do item.
type Task struct {
	ID          int64   `json:"id"`
	UUID        string  `json:"uuid"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	PersonID    *int64  `json:"person_id"`
	DealID      *int64  `json:"deal_id"`
	DueAt       *string `json:"due_at"`
	Priority    string  `json:"priority"`
	Completed   bool    `json:"completed"`
	CompletedAt *string `json:"completed_at"`
	Archived    bool    `json:"-"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateTaskInput holds fields for creating a task.
type CreateTaskInput struct {
	Title       string
	Description *string
	PersonID    *int64
	DealID      *int64
	DueAt       *string
	Priority    string
}

// UpdateTaskInput holds optional fields for updating a task.
type UpdateTaskInput struct {
	Title       *string
	Description *string
	PersonID    *int64
	DealID      *int64
	DueAt       *string
	Priority    *string
}

// TaskFilters holds optional filters for listing tasks.
type TaskFilters struct {
	PersonID         *int64
	DealID           *int64
	Overdue          bool
	IncludeCompleted bool
	Limit            int
}

// Relationship represents a person-to-person link.
type Relationship struct {
	ID              int64   `json:"id"`
	PersonID        int64   `json:"person_id"`
	RelatedPersonID int64   `json:"related_person_id"`
	Type            string  `json:"type"`
	Notes           *string `json:"notes"`
	CreatedAt       string  `json:"created_at"`
}

// Valid entity types for tagging.
var EntityTypes = []string{"person", "organization", "deal", "interaction"}

// ValidEntityType checks if the given type is valid.
func ValidEntityType(t string) bool {
	for _, et := range EntityTypes {
		if et == t {
			return true
		}
	}
	return false
}

// Interaction types
var InteractionTypes = []string{"call", "email", "meeting", "note", "message"}

// Deal stages
var DealStages = []string{"lead", "prospect", "proposal", "negotiation", "won", "lost"}

// ValidDealStage checks if the given stage is valid.
func ValidDealStage(s string) bool {
	for _, ds := range DealStages {
		if ds == s {
			return true
		}
	}
	return false
}

// Priority levels
var Priorities = []string{"low", "medium", "high"}

// ValidPriority checks if the given priority is valid.
func ValidPriority(p string) bool {
	for _, pr := range Priorities {
		if pr == p {
			return true
		}
	}
	return false
}

// Relationship types
var RelationshipTypes = []string{"colleague", "friend", "manager", "mentor", "referred-by"}

// ValidRelationshipType checks if the given type is valid.
func ValidRelationshipType(t string) bool {
	for _, rt := range RelationshipTypes {
		if rt == t {
			return true
		}
	}
	return false
}
