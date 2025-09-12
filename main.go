package vrc_event_notify

import (
	"context"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type EventInfo struct {
	Summary       string `firestore:"channel_id,omitempty"`
	Description   string `firestore:"description,omitempty"`
	StartDateTime string `firestore:"start_date_time,omitempty"`
	EndDateTime   string `firestore:"end_date_time,omitempty"`
}

func Main() error {
	APIKey := os.Getenv("API_KEY")
	calendarID := os.Getenv("CALENDAR_ID")
	projectID := os.Getenv("PROJECT_ID")

	ctx := context.Background()
	calendarService, err := calendar.NewService(ctx, option.WithAPIKey(APIKey))
	if err != nil {
		slog.Error("Failed to create Calendar service:: " + err.Error())
		return err
	}

	client, err := firestore.NewClientWithDatabase(ctx, projectID, projectID)
	if err != nil {
		slog.Error("Failed to create Firestore client: " + err.Error())
		return err
	}
	defer func(client *firestore.Client) {
		err := client.Close()
		if err != nil {
			slog.Error("Failed to close Firestore client: " + err.Error())
		}
	}(client)

	// 24時間以内に更新されたカレンダーイベントを取得するための処理
	updatedMin := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	slog.Info("updatedMin: " + updatedMin)

	list := calendarService.Events.List(calendarID).MaxResults(5).OrderBy("updated").SingleEvents(true).UpdatedMin(updatedMin)
	event, err := list.Do()
	if err != nil {
		slog.Error("failed to list events:: " + err.Error())
		return err
	}

	for _, i := range event.Items {
		slog.Info("event",
			slog.String("event_id", i.Id),
			slog.String("summary", i.Summary),
			slog.String("description", i.Description),
			slog.String("start_date_time", i.Start.DateTime),
			slog.String("end_date_time", i.End.DateTime),
		)
		e := &EventInfo{
			Summary:       i.Summary,
			Description:   i.Description,
			StartDateTime: i.Start.DateTime,
			EndDateTime:   i.End.DateTime,
		}
		_, err := client.Collection("events").Doc(i.Id).Set(ctx, e)
		if err != nil {
			slog.Error("failed to add event: " + err.Error())
		}
	}

	return nil
}
