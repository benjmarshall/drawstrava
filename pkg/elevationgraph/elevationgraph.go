package elevationgraph

import (
	"errors"
	"log"
	"math"
	"time"

	"github.com/fogleman/gg"
	"github.com/strava/go.strava"

	"github.com/benjmarshall/drawstrava/pkg/activities"
)

// Type defines the structure for an elevationgraph object
type Type struct {
	activityList []activities.Activity
	dc           *gg.Context
	mergedStream *mergedStream
}

type mergedStream struct {
	data []float64 // The data stream itself
	tags []tag     // A list of tags identifying where in the merges stream different rides are
}

type tag struct {
	rideName     string
	dataIndex    int // Represents the start index in the merged stream of the ride
	oneDirection bool
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

	activityList, err := strava.NewCurrentAthleteService(client).ListActivities().Before(int(before)).After(int(after)).Do()
	if err != nil {
		return err
	}

	for _, a := range activityList {
		t.activityList = append(t.activityList, activities.New(a))
	}

	return nil
}

func (t *Type) getActivityStreams(accessToken string) error {

	client := strava.NewClient(accessToken)

	for i := range t.activityList {
		err := t.activityList[i].GetStream(client)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Type) mergeStreams() {
	t.mergedStream = new(mergedStream)
	// Add activityList in reverse order as they are stored by date newest first
	// and we want a chronological output
	for i := len(t.activityList) - 1; i >= 0; i-- {
		t.mergedStream.addStream(&t.activityList[i])
	}
}

func (m *mergedStream) addStream(a *activities.Activity) {
	stats := a.GetStats()

	// See if the ride is a downhill/uphill only
	isOneWay := false
	if math.Abs(stats.StartEl-stats.EndEl) > 0.25*stats.MaxEl-stats.MinEl {
		isOneWay = true
	}

	// Get the last elevation value of the merged stream
	var lastEl float64
	if len(m.data) != 0 {
		lastEl = m.data[len(m.data)-1]
	}

	// Pull in the data
	data := a.GetElevationData()

	// Start the new data where the last ride left off.
	var diff float64
	if isOneWay {
		diff = data[len(data)-1] - lastEl
	} else {
		diff = data[0] - lastEl
	}
	for i := range data {
		data[i] -= diff
	}

	m.tags = append(m.tags, tag{rideName: a.GetName(), dataIndex: len(m.data), oneDirection: isOneWay})
	m.data = append(m.data, data...)
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
	tagCount := 0

	for i := 1; i < len(t.mergedStream.data); i++ {
		t.dc.SetRGB(0, 0, 0)
		if tagCount < len(t.mergedStream.tags)-1 {
			if i == t.mergedStream.tags[tagCount+1].dataIndex {
				if t.mergedStream.tags[tagCount+1].oneDirection == true {
					t.dc.SetRGBA(0, 0, 0, 0.5)
				}
				tagCount++
			}
		}
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
