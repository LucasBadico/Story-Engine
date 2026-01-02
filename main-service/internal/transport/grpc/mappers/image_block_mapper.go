package mappers

import (
	"github.com/story-engine/main-service/internal/core/story"
	imageblockpb "github.com/story-engine/main-service/proto/image_block"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ImageBlockToProto converts an image block domain entity to a protobuf message
func ImageBlockToProto(i *story.ImageBlock) *imageblockpb.ImageBlock {
	if i == nil {
		return nil
	}

	var chapterID *string
	if i.ChapterID != nil {
		id := i.ChapterID.String()
		chapterID = &id
	}

	var orderNum *int32
	if i.OrderNum != nil {
		on := int32(*i.OrderNum)
		orderNum = &on
	}

	var altText *string
	if i.AltText != nil {
		at := *i.AltText
		altText = &at
	}

	var caption *string
	if i.Caption != nil {
		c := *i.Caption
		caption = &c
	}

	var width *int32
	if i.Width != nil {
		w := int32(*i.Width)
		width = &w
	}

	var height *int32
	if i.Height != nil {
		h := int32(*i.Height)
		height = &h
	}

	return &imageblockpb.ImageBlock{
		Id:        i.ID.String(),
		ChapterId: chapterID,
		OrderNum:  orderNum,
		Kind:      string(i.Kind),
		ImageUrl:  i.ImageURL,
		AltText:   altText,
		Caption:   caption,
		Width:     width,
		Height:    height,
		CreatedAt: timestamppb.New(i.CreatedAt),
		UpdatedAt: timestamppb.New(i.UpdatedAt),
	}
}

