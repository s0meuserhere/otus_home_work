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

func newTestClient(t *testing.T) gen.EventServiceClient {
	t.Helper()

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

	return gen.NewEventServiceClient(conn)
}

func TestGRPCEvents(t *testing.T) {
	client := newTestClient(t)
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

func TestGRPCUpdateEvent(t *testing.T) {
	client := newTestClient(t)
	userID := uuid.New()
	start := time.Date(2030, 6, 16, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	created, err := client.CreateEvent(context.Background(), &gen.CreateEventRequest{
		Title:       "meeting",
		DateStart:   timestamppb.New(start),
		DateEnd:     timestamppb.New(end),
		Description: "desc",
		UserId:      userID.String(),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newStart := start.Add(2 * time.Hour)
	newEnd := newStart.Add(30 * time.Minute)
	updated, err := client.UpdateEvent(context.Background(), &gen.UpdateEventRequest{
		Id:                 created.GetEvent().GetId(),
		Title:              "updated",
		DateStart:          timestamppb.New(newStart),
		DateEnd:            timestamppb.New(newEnd),
		Description:        "new desc",
		UserId:             userID.String(),
		NotifyShiftSeconds: 60,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	got := updated.GetEvent()
	if got.GetId() != created.GetEvent().GetId() {
		t.Fatalf("id: got %s, want %s", got.GetId(), created.GetEvent().GetId())
	}
	if got.GetTitle() != "updated" || got.GetDescription() != "new desc" || got.GetNotifyShiftSeconds() != 60 {
		t.Fatalf("unexpected event: %+v", got)
	}
	if !got.GetDateStart().AsTime().Equal(newStart) || !got.GetDateEnd().AsTime().Equal(newEnd) {
		t.Fatalf("dates: got %v-%v, want %v-%v",
			got.GetDateStart().AsTime(), got.GetDateEnd().AsTime(), newStart, newEnd)
	}
}

func TestGRPCErrorCodes(t *testing.T) {
	client := newTestClient(t)
	userID := uuid.New()
	start := time.Date(2030, 6, 16, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	existingID := uuid.New()
	_, err := client.CreateEvent(context.Background(), &gen.CreateEventRequest{
		Id:          existingID.String(),
		Title:       "meeting",
		DateStart:   timestamppb.New(start),
		DateEnd:     timestamppb.New(end),
		Description: "desc",
		UserId:      userID.String(),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	tests := []struct {
		name string
		call func() error
		want codes.Code
	}{
		{
			name: "update not found",
			call: func() error {
				_, err := client.UpdateEvent(context.Background(), &gen.UpdateEventRequest{
					Id:          uuid.NewString(),
					Title:       "missing",
					DateStart:   timestamppb.New(start.Add(24 * time.Hour)),
					DateEnd:     timestamppb.New(end.Add(24 * time.Hour)),
					Description: "desc",
					UserId:      uuid.NewString(),
				})
				return err
			},
			want: codes.NotFound,
		},
		{
			name: "delete not found",
			call: func() error {
				_, err := client.DeleteEvent(context.Background(), &gen.DeleteEventRequest{Id: uuid.NewString()})
				return err
			},
			want: codes.NotFound,
		},
		{
			name: "create with existing id",
			call: func() error {
				_, err := client.CreateEvent(context.Background(), &gen.CreateEventRequest{
					Id:          existingID.String(),
					Title:       "duplicate",
					DateStart:   timestamppb.New(start.Add(48 * time.Hour)),
					DateEnd:     timestamppb.New(end.Add(48 * time.Hour)),
					Description: "desc",
					UserId:      uuid.NewString(),
				})
				return err
			},
			want: codes.AlreadyExists,
		},
		{
			name: "create when date is busy",
			call: func() error {
				_, err := client.CreateEvent(context.Background(), &gen.CreateEventRequest{
					Title:       "overlap",
					DateStart:   timestamppb.New(start.Add(30 * time.Minute)),
					DateEnd:     timestamppb.New(end.Add(30 * time.Minute)),
					Description: "desc",
					UserId:      userID.String(),
				})
				return err
			},
			want: codes.AlreadyExists,
		},
		{
			name: "delete with invalid id",
			call: func() error {
				_, err := client.DeleteEvent(context.Background(), &gen.DeleteEventRequest{Id: "not-a-uuid"})
				return err
			},
			want: codes.InvalidArgument,
		},
		{
			name: "list by week not monday",
			call: func() error {
				_, err := client.ListEventsByWeek(context.Background(), &gen.ListEventsRequest{
					Date: timestamppb.New(time.Date(2030, 6, 16, 0, 0, 0, 0, time.UTC)),
				})
				return err
			},
			want: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if status.Code(err) != tt.want {
				t.Fatalf("code: got %v, want %v, err=%v", status.Code(err), tt.want, err)
			}
		})
	}
}
