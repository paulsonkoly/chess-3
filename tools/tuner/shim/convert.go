package shim

import (
	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
)

// epdInfoFromGrpc converts a GRPC response to EPDInfo.
func epdInfoFromGrpc(er *pb.EPDInfo) (EPDInfo, error) {
	result := EPDInfo{Filename: er.Filename}

	chk, err := checksum.FromBytes(er.Checksum)
	if err != nil {
		return result, err
	}
	result.Checksum = chk

	return result, nil
}

// ToGrpc converts an EPDInfo to GRPC representation.
func (e EPDInfo) toGrpc() (*pb.EPDInfo, error) {
	return &pb.EPDInfo{Filename: e.Filename, Checksum: e.Checksum.Bytes()}, nil
}

// jobFromGrpc converts a GRPC JobResponse to Job.
func jobFromGrpc(jr *pb.JobResponse) (Job, error) {
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

// toGrpc converts a job item to GRPC protobuf.
func (j Job) toGrpc() (*pb.JobResponse, error) {
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

// toGrpc converts a Result type to GRPC representation.
func (r Result) toGrpc() (*pb.ResultRequest, error) {
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

// resultFromGrpc converts GRPC representation to Result type.
func resultFromGrpc(rr *pb.ResultRequest) (Result, error) {
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
