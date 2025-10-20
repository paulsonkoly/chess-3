// package shim shimmies data between the application and the GRPC layer hiding
// representational limitations of the protocol from the app.
package shim

import (
	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
)

type EPDInfo struct {
	Filename string            // Filename is the base name of the EPD file.
	Checksum checksum.Checksum // Checksum is the whole file content checksum.
}

// Job is a workload sent to a client.
type Job struct {
	UUID         uuid.UUID         // UUID uniquely identifies a job request/response.
	Epoch        int               // Epoch is the tuning epoch.
	Range        tuning.Range      // Range is the EPD range for the workload.
	Coefficients tuning.Vector     // Coefficients are the tuned coefficients for the current batch.
	Checksum     checksum.Checksum // Checksum is the EPD chunk checksum.
	K            float64           // K is the sigmoid constant.
}

// Result is the workload result.
type Result struct {
	UUID      uuid.UUID     // UUID identifies which workload request this is a result of.
	Gradients tuning.Vector // Gradients is the computed gradient vector.
}
