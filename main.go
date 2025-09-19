package vrc_event_notify

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type EventInfo struct {
	Summary       string `bson:"summary,omitempty"`
	Description   string `bson:"description,omitempty"`
	StartDateTime string `bson:"start_date_time,omitempty"`
	EndDateTime   string `bson:"end_date_time,omitempty"`
}

func Main() error {
	APIKey := os.Getenv("API_KEY")
	calendarID := os.Getenv("CALENDAR_ID")
	uri := os.Getenv("MONGODB_URI")

	ctx := context.Background()
	calendarService, err := calendar.NewService(ctx, option.WithAPIKey(APIKey))
	if err != nil {
		slog.Error("Failed to create Calendar service:: " + err.Error())
		return err
	}

	// Uses the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	// Defines the options for the MongoDB client
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	// Creates a new client and connects to the server
	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// 24時間以内に更新されたカレンダーイベントを取得するための処理
	updatedMin := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	slog.Info("updatedMin: " + updatedMin)

	list := calendarService.Events.List(calendarID).MaxResults(50).OrderBy("updated").SingleEvents(true).UpdatedMin(updatedMin)
	event, err := list.Do()
	if err != nil {
		slog.Error("failed to list events:: " + err.Error())
		return err
	}

	eventInfo := make(map[string]EventInfo)
	for _, i := range event.Items {
		slog.Info("event",
			slog.String("event_id", i.Id),
			slog.String("summary", i.Summary),
			slog.String("description", i.Description),
			slog.String("start_date_time", i.Start.DateTime),
			slog.String("end_date_time", i.End.DateTime),
		)
		eventInfo[i.Id] = EventInfo{
			Summary:       i.Summary,
			Description:   i.Description,
			StartDateTime: i.Start.DateTime,
			EndDateTime:   i.End.DateTime,
		}
	}

	_, err = client.Database("mongo-db").Collection("events").InsertMany(ctx, eventInfo)
	if err != nil {
		slog.Error("failed to add event: " + err.Error())
	}

	return nil
}
