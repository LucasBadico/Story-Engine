//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	pb "github.com/story-engine/main-service/proto/plot"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestPlotHandler_CreateAndGet(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewPlotServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)

	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("create plot", func(t *testing.T) {
		resp, err := client.CreatePlot(ctx, &pb.CreatePlotRequest{
			StoryId:     storyResp.Story.Id,
			Title:       "M-sama's Identity",
			Type:        "main",
			Description: "Who is M-sama and what do they want with Haruka?",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Plot.Title != "M-sama's Identity" {
			t.Errorf("expected title 'M-sama's Identity', got '%s'", resp.Plot.Title)
		}
		if resp.Plot.Status != "open" {
			t.Errorf("expected status 'open', got '%s'", resp.Plot.Status)
		}

		// Get
		getResp, err := client.GetPlot(ctx, &pb.GetPlotRequest{Id: resp.Plot.Id})
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if getResp.Plot.Type != "main" {
			t.Errorf("expected type 'main', got '%s'", getResp.Plot.Type)
		}
	})

	t.Run("get non-existing", func(t *testing.T) {
		_, err := client.GetPlot(ctx, &pb.GetPlotRequest{Id: uuid.New().String()})
		if err == nil {
			t.Fatal("expected error")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestPlotHandler_ListAndUpdate(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewPlotServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)

	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	// Create 2 plots
	plot1, err := client.CreatePlot(ctx, &pb.CreatePlotRequest{
		StoryId: storyResp.Story.Id, Title: "Plot A", Type: "main", Description: "Main plot",
	})
	if err != nil {
		t.Fatalf("failed to create plot 1: %v", err)
	}
	_, err = client.CreatePlot(ctx, &pb.CreatePlotRequest{
		StoryId: storyResp.Story.Id, Title: "Plot B", Type: "secondary", Description: "Secondary",
	})
	if err != nil {
		t.Fatalf("failed to create plot 2: %v", err)
	}

	t.Run("list all", func(t *testing.T) {
		resp, err := client.ListPlotsByStory(ctx, &pb.ListPlotsByStoryRequest{StoryId: storyResp.Story.Id})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Plots) != 2 {
			t.Errorf("expected 2 plots, got %d", len(resp.Plots))
		}
	})

	t.Run("update status to closed", func(t *testing.T) {
		resp, err := client.UpdatePlot(ctx, &pb.UpdatePlotRequest{
			Id:     plot1.Plot.Id,
			Status: "closed",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Plot.Status != "closed" {
			t.Errorf("expected status 'closed', got '%s'", resp.Plot.Status)
		}
	})
}

func TestPlotHandler_Developments(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewPlotServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)

	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	plotResp, err := client.CreatePlot(ctx, &pb.CreatePlotRequest{
		StoryId: storyResp.Story.Id, Title: "The Ribbon", Type: "secondary", Description: "Thread about the crimson ribbon",
	})
	if err != nil {
		t.Fatalf("failed to create plot: %v", err)
	}

	t.Run("create developments", func(t *testing.T) {
		_, err := client.CreatePlotDevelopment(ctx, &pb.CreatePlotDevelopmentRequest{
			PlotId:        plotResp.Plot.Id,
			ChapterNumber: 1,
			Action:        "open",
			Description:   "Ribbon first mentioned",
		})
		if err != nil {
			t.Fatalf("failed to create dev 1: %v", err)
		}

		_, err = client.CreatePlotDevelopment(ctx, &pb.CreatePlotDevelopmentRequest{
			PlotId:             plotResp.Plot.Id,
			ChapterNumber:      3,
			Action:             "develop",
			Description:        "Ribbon appears in Haruka's drawer",
			HasTimelineDay:     true,
			TimelineDay:        -40,
			TimelineConfidence: "explicit",
		})
		if err != nil {
			t.Fatalf("failed to create dev 2: %v", err)
		}

		// List all
		listResp, err := client.ListPlotDevelopments(ctx, &pb.ListPlotDevelopmentsRequest{
			PlotId: plotResp.Plot.Id,
		})
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(listResp.PlotDevelopments) != 2 {
			t.Errorf("expected 2 developments, got %d", len(listResp.PlotDevelopments))
		}

		// List with as_of_chapter=2 → should return 1
		listResp2, err := client.ListPlotDevelopments(ctx, &pb.ListPlotDevelopmentsRequest{
			PlotId:      plotResp.Plot.Id,
			AsOfChapter: 2,
		})
		if err != nil {
			t.Fatalf("failed to list with filter: %v", err)
		}
		if len(listResp2.PlotDevelopments) != 1 {
			t.Errorf("expected 1 development (ch <= 2), got %d", len(listResp2.PlotDevelopments))
		}
	})
}

func TestPlotHandler_Delete(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewPlotServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)

	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	plotResp, err := client.CreatePlot(ctx, &pb.CreatePlotRequest{
		StoryId: storyResp.Story.Id, Title: "To Delete", Type: "minor", Description: "Will be removed",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	_, err = client.DeletePlot(ctx, &pb.DeletePlotRequest{Id: plotResp.Plot.Id})
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	_, err = client.GetPlot(ctx, &pb.GetPlotRequest{Id: plotResp.Plot.Id})
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}
