package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/benjmarshall/drawstrava/pkg/elevationgraph"
)

func main() {
	var activityID int64
	var athleteID int64
	var accessToken string

	flag.Int64Var(&activityID, "activity", 1121238188, "Strava Activity Id")
	flag.Int64Var(&athleteID, "athlete", 491188, "Strava Athlete Id")
	flag.StringVar(&accessToken, "token", "", "Access Token")

	flag.Parse()

	if accessToken == "" {
		accessToken = os.Getenv("STRAVATOKEN")
		if accessToken == "" {
			fmt.Println("\nPlease provide an access_token via cli flag or the 'STRAVATOKEN' environment variable, one can be found at https://www.strava.com/settings/api")

			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	e := elevationgraph.New(athleteID)
	e.MakeImage(accessToken)

}
