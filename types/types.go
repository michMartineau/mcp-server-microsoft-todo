// Package types defines data structures for Microsoft Graph API responses.
package types

import "time"

// TodoTaskList represents a Microsoft To-Do task list.
type TodoTaskList struct {
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	IsOwner       bool   `json:"isOwner"`
	IsShared      bool   `json:"isShared"`
	WellknownName string `json:"wellknownListName,omitempty"`
}

// TodoTaskListsResponse is the API response for listing task lists.
type TodoTaskListsResponse struct {
	Value    []TodoTaskList `json:"value"`
	NextLink string         `json:"@odata.nextLink,omitempty"`
}

// TodoTask represents a single task in Microsoft To-Do.
type TodoTask struct {
	ID                   string          `json:"id,omitempty"`
	Title                string          `json:"title"`
	Body                 *ItemBody       `json:"body,omitempty"`
	Importance           string          `json:"importance,omitempty"`
	Status               string          `json:"status,omitempty"`
	CreatedDateTime      *time.Time      `json:"createdDateTime,omitempty"`
	LastModifiedDateTime *time.Time      `json:"lastModifiedDateTime,omitempty"`
	DueDateTime          *DateTimeZone   `json:"dueDateTime,omitempty"`
	CompletedDateTime    *DateTimeZone   `json:"completedDateTime,omitempty"`
	Recurrence           *PatternedRecurrence `json:"recurrence,omitempty"`
}

// ItemBody represents the body content of a task.
type ItemBody struct {
	Content     string `json:"content,omitempty"`
	ContentType string `json:"contentType,omitempty"` // "text" or "html"
}

// DateTimeZone represents a date/time with timezone info.
type DateTimeZone struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

// PatternedRecurrence defines recurrence pattern for a task.
type PatternedRecurrence struct {
	Pattern *RecurrencePattern `json:"pattern,omitempty"`
	Range   *RecurrenceRange   `json:"range,omitempty"`
}

// RecurrencePattern defines how often a task recurs.
type RecurrencePattern struct {
	Type           string   `json:"type,omitempty"`
	Interval       int      `json:"interval,omitempty"`
	DaysOfWeek     []string `json:"daysOfWeek,omitempty"`
	DayOfMonth     int      `json:"dayOfMonth,omitempty"`
	FirstDayOfWeek string   `json:"firstDayOfWeek,omitempty"`
}

// RecurrenceRange defines the duration of recurrence.
type RecurrenceRange struct {
	Type                string `json:"type,omitempty"`
	StartDate           string `json:"startDate,omitempty"`
	EndDate             string `json:"endDate,omitempty"`
	NumberOfOccurrences int    `json:"numberOfOccurrences,omitempty"`
}

// TodoTasksResponse is the API response for listing tasks.
type TodoTasksResponse struct {
	Value    []TodoTask `json:"value"`
	NextLink string     `json:"@odata.nextLink,omitempty"`
}

// GraphError represents an error from Microsoft Graph API.
type GraphError struct {
	Error GraphErrorDetail `json:"error"`
}

// GraphErrorDetail contains error details.
type GraphErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TokenResponse holds OAuth tokens from Azure AD.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// DeviceCodeResponse is returned when initiating device code flow.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

// StoredTokens is the format for persisted tokens.
type StoredTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}
