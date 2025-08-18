package checks

import "time"

// Severity levels for check results
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// CheckResult represents the result of a single check
type CheckResult struct {
	Name      string
	Severity  Severity
	Message   string
	Details   []string
	Timestamp time.Time
}

// Check interface that all checks must implement
type Check interface {
	Name() string
	Run() CheckResult
	RequiresRoot() bool
}

// Results aggregates all check results
type Results struct {
	checks   []CheckResult
	errors   []string
	warnings []string
	info     []string
}

// NewResults creates a new Results instance
func NewResults() Results {
	return Results{
		checks:   []CheckResult{},
		errors:   []string{},
		warnings: []string{},
		info:     []string{},
	}
}

// AddResult adds a check result to the results
func (r *Results) AddResult(result CheckResult) {
	r.checks = append(r.checks, result)
	
	switch result.Severity {
	case SeverityError, SeverityCritical:
		r.errors = append(r.errors, result.Message)
	case SeverityWarning:
		r.warnings = append(r.warnings, result.Message)
	case SeverityInfo:
		r.info = append(r.info, result.Message)
	}
}

// GetErrors returns all error messages
func (r *Results) GetErrors() []string {
	return r.errors
}

// GetWarnings returns all warning messages
func (r *Results) GetWarnings() []string {
	return r.warnings
}

// GetInfo returns all info messages
func (r *Results) GetInfo() []string {
	return r.info
}

// GetAllChecks returns all check results
func (r *Results) GetAllChecks() []CheckResult {
	return r.checks
}