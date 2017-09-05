package activities

import strava "github.com/strava/go.strava"

// Activity is a wrapper around a strava activity providing both the
// summary and the acitivy streams.
type Activity struct {
	activitySummary *strava.ActivitySummary
	activityStream  *strava.StreamSet
}

// ActivityStats contains useful values extracted from an activities data stream.
type ActivityStats struct {
	MaxEl   float64
	MinEl   float64
	StartEl float64
	EndEl   float64
}

// New creates a new Activity from an activity summary.
// A blank stream set is created which can be populated using 'getStreams'
func New(summary *strava.ActivitySummary) Activity {
	return Activity{activitySummary: summary, activityStream: new(strava.StreamSet)}
}

// GetStream fetches the activity streams for an Activity.
func (a *Activity) GetStream(client *strava.Client) error {
	types := []strava.StreamType{"altitude"}
	resolution := "low"
	seriesType := "distance"

	var err error
	a.activityStream, err = strava.NewActivityStreamsService(client).Get(a.activitySummary.Id, types).Resolution(resolution).SeriesType(seriesType).Do()
	if err != nil {
		return err
	}
	return nil
}

// GetStats returns the ActivityStats struct for an Activity.
func (a *Activity) GetStats() ActivityStats {
	stats := new(ActivityStats)

	stats.StartEl = a.activityStream.Elevation.Data[0]
	stats.EndEl = a.activityStream.Elevation.Data[len(a.activityStream.Elevation.Data)-1]

	for i, val := range a.activityStream.Elevation.Data {
		if val > stats.MaxEl || i == 0 {
			stats.MaxEl = val
		}
		if val < stats.MinEl || i == 0 {
			stats.MinEl = val
		}
	}
	return *stats
}

// GetElevationData returns the data array for the elevation stream.
func (a *Activity) GetElevationData() []float64 {
	return a.activityStream.Elevation.Data
}

// GetName returns the Activity name.
func (a *Activity) GetName() string {
	return a.activitySummary.Name
}
