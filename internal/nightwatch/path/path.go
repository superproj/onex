package path

import "fmt"

const (
	JobDataName         = "data"     // Original data
	JobEmbeddedDataName = "embedded" // Data after embedding
	JobResultName       = "result"   // Data after model evaluation
)

// Path represents a generic path structure with a prefix.
type Path struct {
	Prefix string
}

var (
	Job Path = Path{Prefix: "job"}
)

// Path constructs a complete path for a job.
func (p Path) Path(jobID string, name string) string {
	return fmt.Sprintf("%s/%s/%s.json", p.Prefix, jobID, name)
}
