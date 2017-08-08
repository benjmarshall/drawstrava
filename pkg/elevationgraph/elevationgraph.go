package elevationgraph

import (
	"time"

	"github.com/fogleman/gg"
	"github.com/strava/go.strava"
)

// Type defines the structure for an elevationgraph object
type Type struct {
	activities   []activity
	dc           *gg.Context
	athleteID    int64
	mergedStream []float64
}

type activity struct {
	activitySummary *strava.ActivitySummary
	activiyStream   *strava.StreamSet
}

// New creates a new elevationgraph structure
func New(athleteID int64) *Type {
	w := 1024
	h := 300

	t := new(Type)
	t.dc = gg.NewContext(w, h)
	t.athleteID = athleteID
	return t
}

// MakeImage create a png image of the elevationgraph type
func (t *Type) MakeImage(accessToken string) {
	t.getActivities(accessToken)
	t.mergeStreams()
	t.drawImage()
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

	types := []strava.StreamType{"altitude"}
	resolution := "low"
	seriesType := "distance"

	for _, a := range activities {

		s, err := strava.NewActivityStreamsService(client).Get(a.Id, types).Resolution(resolution).SeriesType(seriesType).Do()
		if err != nil {
			return err
		}

		t.activities = append(t.activities, activity{activitySummary: a, activiyStream: s})
	}
	return nil
}

func (t *Type) mergeStreams() {
	for _, a := range t.activities {
		t.mergedStream = append(t.mergedStream, a.activiyStream.Elevation.Data...)
	}
}

func (t *Type) drawImage() {
	t.dc.SetRGB(1, 1, 1)
	t.dc.Clear()
	t.dc.SetRGB(0, 0, 0)
	t.dc.SetLineWidth(2.0)
	incr := float64(t.dc.Width()) / float64(len(t.mergedStream))
	minVal := minFloatInSlice(t.mergedStream)
	hRange := maxFloatInSlice(t.mergedStream) - minVal
	hScale := float64(t.dc.Height()) / hRange
	x1 := 0.0

	for i := 1; i < len(t.mergedStream); i++ {
		y1 := float64(t.dc.Height()) - ((t.mergedStream[i-1] - minVal) * hScale)
		y2 := float64(t.dc.Height()) - ((t.mergedStream[i] - minVal) * hScale)
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
