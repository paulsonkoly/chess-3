// package shim shimmies data between the application and the GRPC layer hiding
// representational limitations of the protocol from the app.
package shim

import (
	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
)

type EPDInfo struct {
	Filename string            // Filename is the base name of the EPD file.
	Checksum checksum.Checksum // Checksum is the whole file content checksum.
}

// ToGrpc converts an EPDInfo to GRPC representation.
func (e EPDInfo) ToGrpc() (*pb.EPDInfo, error) {
	return &pb.EPDInfo{Filename: e.Filename, Checksum: e.Checksum.Bytes()}, nil
}

// EPDInfoFromGrpc converts a GRPC response to EPDInfo.
func EPDInfoFromGrpc(er *pb.EPDInfo) (EPDInfo, error) {
	result := EPDInfo{Filename: er.Filename}

	chk, err := checksum.FromBytes(er.Checksum)
	if err != nil {
		return result, err
	}
	result.Checksum = chk

	return result, nil
}

// Job is a workload sent to a client.
type Job struct {
	UUID         uuid.UUID         // UUID uniquely identifies a job request/response.
	Epoch        int               // Epoch is the tuning epoch.
	Range        tuning.Range      // Range is the EPD range for the workload.
	Coefficients *tuning.Coeffs    // Coefficients are the tuned coefficients for the current batch.
	Checksum     checksum.Checksum // Checksum is the EPD chunk checksum.
	K            float64           // K is the sigmoid constant.
}

// ToGrpc converts a job item to GRPC protobuf.
func (j Job) ToGrpc() (*pb.JobResponse, error) {
	uuidBytes, err := j.UUID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	floats, err := j.Coefficients.Floats(tuning.DefaultTargets)
	if err != nil {
		return nil, err
	}

	return &pb.JobResponse{
		Uuid:         uuidBytes,
		Epoch:        int32(j.Epoch),
		Start:        int32(j.Range.Start),
		End:          int32(j.Range.End),
		Coefficients: floats,
		Checksum:     j.Checksum.Bytes(),
		K:            j.K,
	}, nil
}

// JobFromGrpc converts a GRPC JobResponse to Job.
func JobFromGrpc(jr *pb.JobResponse) (Job, error) {
	result := Job{}

	coeffs, err := tuning.EngineCoeffs()
	if err != nil {
		return result, err
	}

	err = coeffs.SetFloats(tuning.DefaultTargets, jr.Coefficients)
	if err != nil {
		return result, err
	}
	result.Coefficients = &coeffs

	result.UUID, err = uuid.FromBytes(jr.Uuid)
	if err != nil {
		return result, err
	}

	result.Checksum, err = checksum.FromBytes(jr.Checksum)
	if err != nil {
		return result, err
	}

	result.K = jr.K
	result.Epoch = int(jr.Epoch)
	result.Range = tuning.Range{Start: int(jr.Start), End: int(jr.End)}

	return result, nil
}

// Result is the workload result.
type Result struct {
	UUID      uuid.UUID      // UUID identifies which workload request this is a result of.
	Gradients *tuning.Coeffs // Gradients is the computed gradient vector.
}

// ToGrpc converts a Result type to GRPC representation.
func (r Result) ToGrpc() (*pb.ResultRequest, error) {
	uuidBytes, err := r.UUID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	floats, err := r.Gradients.Floats(tuning.DefaultTargets)
	if err != nil {
		return nil, err
	}

	return &pb.ResultRequest{Uuid: uuidBytes, Gradients: floats}, nil
}

// ResultFromGrpc converts GRPC representation to Result type.
func ResultFromGrpc(rr *pb.ResultRequest) (Result, error) {
	result := Result{}

	grads := tuning.Coeffs{}

	err := grads.SetFloats(tuning.DefaultTargets, rr.Gradients)
	if err != nil {
		return result, err
	}
	result.Gradients = &grads

	result.UUID, err = uuid.FromBytes(rr.Uuid)
	if err != nil {
		return result, err
	}

	return result, nil
}
