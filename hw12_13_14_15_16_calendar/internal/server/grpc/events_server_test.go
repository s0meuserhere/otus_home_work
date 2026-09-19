package internalgrpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	eventrepo "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/repository/event"
	internalgrpc "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc/gen"
	eventservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/event"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGRPCEvents(t *testing.T) {
	const bufSize = 1024 * 1024

	logg := logger.New("error", "local")
	svc := eventservice.New(eventrepo.NewMemory())
	server := internalgrpc.NewServer(logg, "bufnet", svc)

	lis := bufconn.Listen(bufSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.ServeListener(ctx, lis)
	}()
	t.Cleanup(func() {
		cancel()
		_ = server.Stop(context.Background())
	})

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	client := gen.NewEventServiceClient(conn)
	userID := uuid.New()
	day := time.Date(2030, 6, 16, 0, 0, 0, 0, time.UTC)
	start := day.Add(10 * time.Hour)
	end := start.Add(time.Hour)

	created, err := client.CreateEvent(context.Background(), &gen.CreateEventRequest{
		Title:              "meeting",
		DateStart:          timestamppb.New(start),
		DateEnd:            timestamppb.New(end),
		Description:        "desc",
		UserId:             userID.String(),
		NotifyShiftSeconds: 0,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.GetEvent().GetId() == "" {
		t.Fatal("expected event id")
	}

	list, err := client.ListEventsByDay(context.Background(), &gen.ListEventsRequest{
		Date: timestamppb.New(day),
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list.GetEvents()) != 1 {
		t.Fatalf("expected 1 event, got %d", len(list.GetEvents()))
	}

	_, err = client.DeleteEvent(context.Background(), &gen.DeleteEventRequest{
		Id: created.GetEvent().GetId(),
	})
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = client.CreateEvent(context.Background(), &gen.CreateEventRequest{
		Title:       "",
		DateStart:   timestamppb.New(start),
		DateEnd:     timestamppb.New(end),
		Description: "desc",
		UserId:      userID.String(),
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code: got %v, want InvalidArgument, err=%v", status.Code(err), err)
	}
}
