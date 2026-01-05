package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	worldpb "github.com/story-engine/main-service/proto/world"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimeConfigToProto converts a time config domain entity to a protobuf message
func TimeConfigToProto(tc *world.TimeConfig) *worldpb.TimeConfig {
	if tc == nil {
		return nil
	}

	pb := &worldpb.TimeConfig{
		BaseUnit:      tc.BaseUnit,
		HoursPerDay:   tc.HoursPerDay,
		DaysPerWeek:   int32(tc.DaysPerWeek),
		DaysPerYear:   int32(tc.DaysPerYear),
		MonthsPerYear: int32(tc.MonthsPerYear),
		MonthLengths:  make([]int32, len(tc.MonthLengths)),
		MonthNames:     tc.MonthNames,
		DayNames:       tc.DayNames,
		EraName:        tc.EraName,
		YearZero:       int32(tc.YearZero),
	}

	for i, length := range tc.MonthLengths {
		pb.MonthLengths[i] = int32(length)
	}

	return pb
}

// WorldToProto converts a world domain entity to a protobuf message
func WorldToProto(w *world.World) *worldpb.World {
	if w == nil {
		return nil
	}

	pb := &worldpb.World{
		Id:          w.ID.String(),
		TenantId:    w.TenantID.String(),
		Name:        w.Name,
		Description: w.Description,
		Genre:       w.Genre,
		IsImplicit:  w.IsImplicit,
		CreatedAt:   timestamppb.New(w.CreatedAt),
		UpdatedAt:   timestamppb.New(w.UpdatedAt),
	}

	if w.TimeConfig != nil {
		pb.TimeConfig = TimeConfigToProto(w.TimeConfig)
	}

	return pb
}


