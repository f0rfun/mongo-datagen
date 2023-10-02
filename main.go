package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Data struct {
	Type string
	Data json.RawMessage
}

type PressurePoint struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	// MeasuredAt   time.Time `json:"measuredAt"`
	// SentAt       time.Time `json:"sentAt"`
	Value        float64 `json:"value,string"`
	AlertLow     float64 `json:"alertLow,string"`
	CriticalLow  float64 `json:"criticalLow,string"`
	OorLow       float64 `json:"oorLow,string"`
	AlertHigh    float64 `json:"alertHigh,string"`
	CriticalHigh float64 `json:"criticalHigh,string"`
	OorHigh      float64 `json:"oorHigh,string"`
}

func main() {

	datajson, err := os.ReadFile("testdata/cop_format.json")
	if err != nil {
		log.Fatal(err)
	}

	var data Data
	err = json.Unmarshal(datajson, &data)
	if err != nil {
		log.Fatal(err)
	}

	switch data.Type {
	case "pressure":
		points := make([]PressurePoint, 1)
		err = json.Unmarshal(data.Data, &points)
		if err != nil {
			log.Fatal(err)
		}
		point := fmt.Sprintf(`
		id: %s
		description: %s
		value: %f
		alertLow: %f
		criticalLow: %f
		oorLow: %f
		alertHigh: %f
		criticalHigh: %f
		oorHigh: %f
		`, points[0].ID,
			points[0].Description,
			points[0].Value,
			points[0].AlertLow,
			points[0].CriticalLow,
			points[0].OorLow,
			points[0].AlertHigh,
			points[0].CriticalHigh,
			points[0].OorHigh,
		)
		fmt.Println(point)
		// temperature etc
	}
}
