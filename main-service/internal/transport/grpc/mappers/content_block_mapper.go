package mappers

import (
	"encoding/json"

	"github.com/story-engine/main-service/internal/core/story"
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ContentBlockToProto converts a content block domain entity to a protobuf message
func ContentBlockToProto(c *story.ContentBlock) *contentblockpb.ContentBlock {
	if c == nil {
		return nil
	}

	pb := &contentblockpb.ContentBlock{
		Id:        c.ID.String(),
		Type:      string(c.Type),
		Kind:      string(c.Kind),
		Content:   c.Content,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}

	if c.ChapterID != nil {
		chapterIDStr := c.ChapterID.String()
		pb.ChapterId = &chapterIDStr
	}
	if c.OrderNum != nil {
		orderNum := int32(*c.OrderNum)
		pb.OrderNum = &orderNum
	}

	// Convert metadata to Struct
	metadataStruct, err := contentMetadataToStruct(c.Metadata)
	if err == nil {
		pb.Metadata = metadataStruct
	}

	return pb
}

// ContentAnchorToProto converts a content anchor domain entity to a protobuf message
func ContentAnchorToProto(anchor *story.ContentAnchor) *contentblockpb.ContentAnchor {
	if anchor == nil {
		return nil
	}

	return &contentblockpb.ContentAnchor{
		Id:             anchor.ID.String(),
		ContentBlockId: anchor.ContentBlockID.String(),
		EntityType:     string(anchor.EntityType),
		EntityId:       anchor.EntityID.String(),
		CreatedAt:      timestamppb.New(anchor.CreatedAt),
	}
}

// contentMetadataToStruct converts ContentMetadata to protobuf Struct
func contentMetadataToStruct(metadata story.ContentMetadata) (*structpb.Struct, error) {
	// Convert to JSON first, then to Struct
	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		return nil, err
	}

	return structpb.NewStruct(jsonMap)
}

// ContentMetadataFromStruct converts protobuf Struct to ContentMetadata
func ContentMetadataFromStruct(s *structpb.Struct) (story.ContentMetadata, error) {
	var metadata story.ContentMetadata

	if s == nil {
		return metadata, nil
	}

	// Convert Struct to JSON
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		return metadata, err
	}

	// Unmarshal to ContentMetadata
	if err := json.Unmarshal(jsonBytes, &metadata); err != nil {
		return metadata, err
	}

	return metadata, nil
}

