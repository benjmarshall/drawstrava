package elevationgraph

import (
	"errors"
	"log"
	"time"

	"github.com/fogleman/gg"
	"github.com/strava/go.strava"
)

// Type defines the structure for an elevationgraph object
type Type struct {
	activities   []activity
	dc           *gg.Context
	mergedStream *mergedStream
}

type activity struct {
	activitySummary *strava.ActivitySummary
	activiyStream   *strava.StreamSet
}

type mergedStream struct {
	data []float64 // The data stream itself
	tags []tag     // A list of tags identifying where in the merges stream different rides are
}

type tag struct {
	rideName  string
	dataIndex int // Represents the start index in the merged stream of the ride
}

// New creates a new elevationgraph structure
func New() *Type {
	w := 1500
	h := 300

	t := new(Type)
	t.dc = gg.NewContext(w, h)
	return t
}

// MakeImage create a png image of the elevationgraph type
func (t *Type) MakeImage(accessToken string) error {
	err := t.getActivities(accessToken)
	if err != nil {
		log.Println(err)
		return errors.New("error retreiving activity summaries")
	}
	err = t.getActivityStreams(accessToken)
	if err != nil {
		return errors.New("error retreiving activity stream")
	}
	t.mergeStreams()
	t.drawImage()
	return nil
}

func (t *Type) getActivities(accessToken string) error {

	client := strava.NewClient(accessToken)

	timeNow := time.Now()
	before := timeNow.Unix()
	after := timeNow.Add(-1 * time.Hour * 24 * 14).Unix()

	activities, err := strava.NewCurrentAthleteService(client).ListActivities().Before(int(before)).After(int(after)).Do()
	if err != nil {
		return err
	}

	for _, a := range activities {
		t.activities = append(t.activities, activity{activitySummary: a, activiyStream: new(strava.StreamSet)})
	}

	return nil
}

func (t *Type) getActivityStreams(accessToken string) error {

	client := strava.NewClient(accessToken)

	types := []strava.StreamType{"altitude"}
	resolution := "low"
	seriesType := "distance"

	for _, a := range t.activities {
		s, err := strava.NewActivityStreamsService(client).Get(a.activitySummary.Id, types).Resolution(resolution).SeriesType(seriesType).Do()
		if err != nil {
			return err
		}
		*a.activiyStream = *s
	}

	return nil
}

func (t *Type) mergeStreams() {
	t.mergedStream = new(mergedStream)
	// Add activities in reverse order as they are stored by date newest first
	// and we want a chronological output
	for i := len(t.activities) - 1; i >= 0; i-- {
		t.mergedStream.data = append(t.mergedStream.data, t.activities[i].activiyStream.Elevation.Data...)
	}
}

func (t *Type) drawImage() {
	t.dc.SetRGB(1, 1, 1)
	t.dc.Clear()
	t.dc.SetRGB(0, 0, 0)
	t.dc.SetLineWidth(2.0)
	incr := float64(t.dc.Width()) / float64(len(t.mergedStream.data))
	minVal := minFloatInSlice(t.mergedStream.data)
	hRange := maxFloatInSlice(t.mergedStream.data) - minVal
	hScale := float64(t.dc.Height()) / hRange
	x1 := 0.0

	for i := 1; i < len(t.mergedStream.data); i++ {
		y1 := float64(t.dc.Height()) - ((t.mergedStream.data[i-1] - minVal) * hScale)
		y2 := float64(t.dc.Height()) - ((t.mergedStream.data[i] - minVal) * hScale)
		x2 := x1 + incr
		t.dc.DrawLine(x1, y1, x2, y2)
		t.dc.Stroke()
		x1 += incr
	}
	t.dc.SavePNG("out.png")
}

func maxFloatInSlice(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, e := range v {
		if e > m {
			m = e
		}
	}
	return m
}

func minFloatInSlice(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, e := range v {
		if e < m {
			m = e
		}
	}
	return m
}
