package notify

// : ) : ( ical feed output — serves a channel's schedule as a subscribable .ics feed
// any calendar app (google, apple, outlook) can subscribe to this URL

import (
	"fmt"
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/go-chi/chi/v5"
	"radioooooo/internal/episode"
)

type ICalFeedHandler struct {
	episodes *episode.Store
}

func NewICalFeedHandler(episodes *episode.Store) *ICalFeedHandler {
	return &ICalFeedHandler{episodes: episodes}
}

// . ݁₊ ✶ public route — no auth
func (h *ICalFeedHandler) Register(r chi.Router) {
	r.Get("/channels/{id}/schedule.ics", h.serve)
}

func (h *ICalFeedHandler) serve(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")

	start := time.Now().AddDate(0, 0, -90)
	end := time.Now().AddDate(0, 0, 90)

	episodes, err := h.episodes.ListRange(r.Context(), channelID, start, end)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//radiooo//schedule//EN")
	cal.SetName("radiooo schedule")

	for _, ep := range episodes {
		event := cal.AddEvent(ep.ID)
		event.SetCreatedTime(ep.CreatedAt)
		event.SetModifiedAt(ep.UpdatedAt)
		event.SetStartAt(ep.StartTime)
		event.SetEndAt(ep.EndTime)
		event.SetSummary(ep.Title)

		if ep.Description != "" {
			event.SetDescription(ep.Description)
		}
		if ep.Color != nil {
			event.SetProperty(ics.ComponentProperty("COLOR"), *ep.Color)
		}

		status := "CONFIRMED"
		if ep.AutoFilled {
			status = "TENTATIVE"
		}
		event.SetStatus(ics.ObjectStatus(status))
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"schedule-%s.ics\"", channelID))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, cal.Serialize())
}
