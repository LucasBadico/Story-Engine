package handlers

import (
	"context"

	"github.com/google/uuid"
	imageblockapp "github.com/story-engine/main-service/internal/application/story/image_block"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	imageblockpb "github.com/story-engine/main-service/proto/image_block"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImageBlockHandler implements the ImageBlockService gRPC service
type ImageBlockHandler struct {
	imageblockpb.UnimplementedImageBlockServiceServer
	createImageBlockUseCase *imageblockapp.CreateImageBlockUseCase
	getImageBlockUseCase    *imageblockapp.GetImageBlockUseCase
	listImageBlocksUseCase  *imageblockapp.ListImageBlocksUseCase
	updateImageBlockUseCase *imageblockapp.UpdateImageBlockUseCase
	deleteImageBlockUseCase *imageblockapp.DeleteImageBlockUseCase
	logger                  logger.Logger
}

// NewImageBlockHandler creates a new ImageBlockHandler
func NewImageBlockHandler(
	createImageBlockUseCase *imageblockapp.CreateImageBlockUseCase,
	getImageBlockUseCase *imageblockapp.GetImageBlockUseCase,
	listImageBlocksUseCase *imageblockapp.ListImageBlocksUseCase,
	updateImageBlockUseCase *imageblockapp.UpdateImageBlockUseCase,
	deleteImageBlockUseCase *imageblockapp.DeleteImageBlockUseCase,
	logger logger.Logger,
) *ImageBlockHandler {
	return &ImageBlockHandler{
		createImageBlockUseCase: createImageBlockUseCase,
		getImageBlockUseCase:    getImageBlockUseCase,
		listImageBlocksUseCase:  listImageBlocksUseCase,
		updateImageBlockUseCase: updateImageBlockUseCase,
		deleteImageBlockUseCase: deleteImageBlockUseCase,
		logger:                  logger,
	}
}

// CreateImageBlock creates a new image block
func (h *ImageBlockHandler) CreateImageBlock(ctx context.Context, req *imageblockpb.CreateImageBlockRequest) (*imageblockpb.CreateImageBlockResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	if req.ImageUrl == "" {
		return nil, status.Errorf(codes.InvalidArgument, "image_url is required")
	}

	var chapterID *uuid.UUID
	if req.ChapterId != nil && *req.ChapterId != "" {
		cid, err := uuid.Parse(*req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		chapterID = &cid
	}

	var orderNum *int
	if req.OrderNum != nil && *req.OrderNum > 0 {
		on := int(*req.OrderNum)
		orderNum = &on
	}

	kind := story.ImageKind(req.Kind)

	var altText *string
	if req.AltText != nil {
		altText = req.AltText
	}

	var caption *string
	if req.Caption != nil {
		caption = req.Caption
	}

	var width *int
	if req.Width != nil && *req.Width > 0 {
		w := int(*req.Width)
		width = &w
	}

	var height *int
	if req.Height != nil && *req.Height > 0 {
		h := int(*req.Height)
		height = &h
	}

	input := imageblockapp.CreateImageBlockInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
		OrderNum:  orderNum,
		Kind:      kind,
		ImageURL:  req.ImageUrl,
		AltText:   altText,
		Caption:   caption,
		Width:     width,
		Height:    height,
	}

	output, err := h.createImageBlockUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &imageblockpb.CreateImageBlockResponse{
		ImageBlock: mappers.ImageBlockToProto(output.ImageBlock),
	}, nil
}

// GetImageBlock retrieves an image block by ID
func (h *ImageBlockHandler) GetImageBlock(ctx context.Context, req *imageblockpb.GetImageBlockRequest) (*imageblockpb.GetImageBlockResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid image_block id: %v", err)
	}

	output, err := h.getImageBlockUseCase.Execute(ctx, imageblockapp.GetImageBlockInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &imageblockpb.GetImageBlockResponse{
		ImageBlock: mappers.ImageBlockToProto(output.ImageBlock),
	}, nil
}

// ListImageBlocks lists image blocks
func (h *ImageBlockHandler) ListImageBlocks(ctx context.Context, req *imageblockpb.ListImageBlocksRequest) (*imageblockpb.ListImageBlocksResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	var chapterID *uuid.UUID
	if req.ChapterId != nil && *req.ChapterId != "" {
		cid, err := uuid.Parse(*req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		chapterID = &cid
	}

	var chapterIDParam uuid.UUID
	if chapterID != nil {
		chapterIDParam = *chapterID
	}
	output, err := h.listImageBlocksUseCase.Execute(ctx, imageblockapp.ListImageBlocksInput{
		TenantID:  tenantUUID,
		ChapterID: chapterIDParam,
	})
	if err != nil {
		return nil, err
	}

	imageBlocks := make([]*imageblockpb.ImageBlock, len(output.ImageBlocks))
	for i, ib := range output.ImageBlocks {
		imageBlocks[i] = mappers.ImageBlockToProto(ib)
	}

	return &imageblockpb.ListImageBlocksResponse{
		ImageBlocks: imageBlocks,
		TotalCount:  int32(len(output.ImageBlocks)),
	}, nil
}

// UpdateImageBlock updates an existing image block
func (h *ImageBlockHandler) UpdateImageBlock(ctx context.Context, req *imageblockpb.UpdateImageBlockRequest) (*imageblockpb.UpdateImageBlockResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid image_block id: %v", err)
	}

	var imageURL *string
	if req.ImageUrl != nil {
		imageURL = req.ImageUrl
	}

	var altText *string
	if req.AltText != nil {
		altText = req.AltText
	}

	var caption *string
	if req.Caption != nil {
		caption = req.Caption
	}

	var width *int
	if req.Width != nil && *req.Width > 0 {
		w := int(*req.Width)
		width = &w
	}

	var height *int
	if req.Height != nil && *req.Height > 0 {
		h := int(*req.Height)
		height = &h
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := imageblockapp.UpdateImageBlockInput{
		TenantID: tenantUUID,
		ID:       id,
		ImageURL: imageURL,
		AltText:  altText,
		Caption:  caption,
		Width:    width,
		Height:   height,
	}

	output, err := h.updateImageBlockUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &imageblockpb.UpdateImageBlockResponse{
		ImageBlock: mappers.ImageBlockToProto(output.ImageBlock),
	}, nil
}

// DeleteImageBlock deletes an image block
func (h *ImageBlockHandler) DeleteImageBlock(ctx context.Context, req *imageblockpb.DeleteImageBlockRequest) (*imageblockpb.DeleteImageBlockResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid image_block id: %v", err)
	}

	err = h.deleteImageBlockUseCase.Execute(ctx, imageblockapp.DeleteImageBlockInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &imageblockpb.DeleteImageBlockResponse{}, nil
}

