package diagnose

import "github.com/debian-doctor/debian-doctor/internal/fixes"

// Diagnosis represents the result of diagnosing an issue
type Diagnosis struct {
	Issue    string
	Findings []string
	Fixes    []*fixes.Fix
}

// DiagnosisResult contains both the diagnosis and execution status
type DiagnosisResult struct {
	Diagnosis *Diagnosis
	FixExecuted bool
	ExecutionError error
}