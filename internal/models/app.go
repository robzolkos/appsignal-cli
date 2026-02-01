package models

// App represents an AppSignal application
type App struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Organization represents an AppSignal organization
type Organization struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Apps []App  `json:"apps"`
}

// AppList represents a list of apps grouped by organization
type AppList struct {
	Organizations []Organization `json:"organizations"`
}
