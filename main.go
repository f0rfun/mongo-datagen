package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type AssetStatus string

const (
	Active   AssetStatus = "ACTIVE"
	Inactive AssetStatus = "INACTIVE"
)

type Phases string

const (
	Red    Phases = "RED"
	Yellow Phases = "YELLOW"
	Blue   Phases = "BLUE"
)

type Locations string

const (
	Source Locations = "SOURCE"
	Target Locations = "TARGET"
)

type Sensor struct {
	SensorID        int
	Phase           Phases
	Location        Locations
	AssetMetrictype string
}

type Section struct {
	SectionNumber int
	Sensors       []Sensor
}

type Substation struct {
	AdwhID int
	Name   string
}

type Cables struct {
	CircuitID        int
	CircuitName      string
	CircuitVoltage   string
	Status           AssetStatus
	SourceSubstation Substation
	TargetSubstation Substation
	CircuitType      string
	FeederNumber     int
	NoOfSections     int
	Sections         []Section
}

type AssetType string

const (
	Cable AssetType = "CABLE"
	RTU   AssetType = "RTU"
	Gauge AssetType = "GAUGE"
)

type Data struct {
	AssetType          AssetType       `json:"type"`
	RawPressureData    json.RawMessage `json:"pressure"`
	RawTemperatureData json.RawMessage `json:"temperature"`
}

type PressurePoint struct {
	RawSensorID  int     `json:"sensorID"`
	Description  string  `json:"description"`
	MeasuredAt   string  `json:"measuredAt"`
	SentAt       string  `json:"sentAt"`
	Value        float64 `json:"value"`
	AlertLow     float64 `json:"alertLow"`
	CriticalLow  float64 `json:"criticalLow"`
	OorLow       float64 `json:"oorLow"`
	AlertHigh    float64 `json:"alertHigh"`
	CriticalHigh float64 `json:"criticalHigh"`
	OorHigh      float64 `json:"oorHigh"`
}

func main() {

	// SEED RANDOM CABLE DATA

	TOTAL_CIRCUITS := 65
	influxConnString := "http://localhost:8086"
	influxOrgString := "sample-org"
	influxBucketString := "sample-bucket"
	influxTokenString := os.Getenv("INFLUXDB_TOKEN") //exported in command line as `export INFLUXDB_TOKEN=jmFgnVqZjozRkDki3ToANuNGaeTKdpyvSis4cyK3Bs6wlqwk-L_NMJLJHcaqNn221zY26Z-gEMwADBgoeBEq0g==`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI("mongodb://jhwu:secret@localhost:27017/"),
	)

	defer func() {
		cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Fatalf("mongodb disconnect error : %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("connection error :%v", err)
		return
	}

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("ping mongodb error :%v", err)
		return
	}
	fmt.Println("ping success")

	// database and collection
	database := mongoClient.Database("asset")
	cableCollection := database.Collection("cables")
	cableCollection.Drop(ctx)

	status := []AssetStatus{Active, Inactive}
	phaseLocations := []string{"RED_SOURCE", "RED_TARGET", "YELLOW_SOURCE", "YELLOW_TARGET", "BLUE_SOURCE", "BLUE_TARGET"}
	var currsensorid int

	// Generate and insert random asset data
	for i := 1; i <= TOTAL_CIRCUITS; i++ {
		randomStatusIndex := rand.Intn(len(status))
		ss := Substation{
			AdwhID: int(rand.Intn(65)),
			Name:   randomString(3),
		}
		ts := Substation{
			AdwhID: int(rand.Intn(65)),
			Name:   randomString(3),
		}

		cable := Cables{
			CircuitID:        int(rand.Intn(65)),
			CircuitName:      randomString(8),
			CircuitVoltage:   randomString(3),
			Status:           status[randomStatusIndex],
			SourceSubstation: ss,
			TargetSubstation: ts,
			CircuitType:      "OIL-FILLED",
			FeederNumber:     int(rand.Intn(5)),
			NoOfSections:     8,
		}

		for j := 1; j <= cable.NoOfSections; j++ {
			var section Section
			section.SectionNumber = j
			var phase Sensor

			for k := 0; k < len(phaseLocations); k++ {
				currsensorid++
				phase = Sensor{
					SensorID:        currsensorid,
					Phase:           "RED",
					Location:        "SOURCE",
					AssetMetrictype: "CABLE_PRESSURE",
				}

				section.Sensors = append(section.Sensors, phase)
			}

			phase = Sensor{
				SensorID:        currsensorid,
				AssetMetrictype: "RTU_TEMPERATURE",
			}

			section.Sensors = append(section.Sensors, phase)

			phase = Sensor{
				SensorID:        currsensorid,
				AssetMetrictype: "RTU_VOLTAGE",
			}

			section.Sensors = append(section.Sensors, phase)

			phase = Sensor{
				SensorID:        currsensorid,
				AssetMetrictype: "GAUGE_VOLTAGE",
			}

			section.Sensors = append(section.Sensors, phase)

			cable.Sections = append(cable.Sections, section)
		}

		_, err := cableCollection.InsertOne(ctx, cable)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Data inserted successfully.")

	//influxdb
	influxClient := influxdb2.NewClientWithOptions(influxConnString, influxTokenString,
		influxdb2.DefaultOptions().SetBatchSize(50),
	)
	writeAPI := influxClient.WriteAPIBlocking(influxOrgString, influxBucketString)
	// queryAPI := influxClient.QueryAPI(influxOrgString)

	//	RETRIEVE RAW SENSOR DATA

	testData, err := os.ReadFile("testdata/cop_format.json")
	if err != nil {
		log.Fatal(err)
	}

	var data Data
	err = json.Unmarshal(testData, &data)
	if err != nil {
		log.Fatal(err)
	}

	SensorDataMap := make(map[int]PressurePoint)
	switch data.AssetType {
	case Cable:
		fmt.Println("=========== Cable sensor data received ===========")
		var rawDataArray []PressurePoint
		err = json.Unmarshal(data.RawPressureData, &rawDataArray)
		if err != nil {
			log.Fatal(err)
		}

		for _, point := range rawDataArray {
			SensorDataMap[point.RawSensorID] = point
		}
	}

	// RETRIEVE CABLE COLLECTION

	fmt.Println("=========== query specific data =====================")
	filter := bson.M{"status": bson.M{"$eq": "active"}}

	// Decode unmarshals BSON back into a user-defined struct
	cursor, err := cableCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatalf("filter query data error : %v", err)
		// return
	}

	for cursor.Next(context.TODO()) {
		var cable Cables
		err = cursor.Decode(&cable)
		if err != nil {
			log.Fatal(err)
		}
		for _, section := range cable.Sections {
			for _, sensor := range section.Sensors {
				point, ok := SensorDataMap[sensor.SensorID]
				if ok {
					tags := map[string]string{
						"cableID":  fmt.Sprint(cable.CircuitID),
						"status":   string(cable.Status),
						"section":  fmt.Sprint(section.SectionNumber),
						"phase":    "RED",
						"location": "SOURCE",
					}
					fields := map[string]interface{}{
						"value":     point.Value,
						"alertHigh": point.AlertHigh,
						"alertLow":  point.AlertLow,
					}
					p := influxdb2.NewPoint("sensor_data", tags, fields, time.Now())
					err = writeAPI.WritePoint(context.TODO(), p)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
		writeAPI.Flush(context.TODO())
	}

	// for sensorID := 0; sensorID < 5; sensorID++ {
	// 	tags := tagsMap[sensorID]
	// 	fields := map[string]interface{}{
	// 		"value":     float64(rand.Int63n(110)),
	// 		"alertLow":  float64(rand.Int63n(100)),
	// 		"alertHigh": float64(rand.Int63n(100)),
	// 	}
	// 	point := write.NewPoint("sensor_data", tags, fields, time.Now())
	// 	time.Sleep(1 * time.Second) // separate points by 1 second

	// 	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// 	// // Can adjust based on sensorID
	// 	// params := tagsMap[0]

	// 	// query := fmt.Sprintf(`from(bucket:"example-bucket")
	// 	// 	|> range(start: -1h)
	// 	// 	|> filter(fn: (r) => r.type == "%s")
	// 	// 	|> filter(fn: (r) => r.cableNo == "%s")
	// 	// 	|> filter(fn: (r) => r.phase == "%s")
	// 	// 	|> filter(fn: (r) => r._field == "value")`,
	// 	// 	params["type"], params["cableNo"], params["phase"])

	// 	// // Get QueryTableResult
	// 	// result, err := queryAPI.QueryWithParams(context.Background(), query, params)
	// 	// if err == nil {
	// 	// 	// Iterate over query response
	// 	// 	for result.Next() {
	// 	// 		// Notice when group key has changed
	// 	// 		if result.TableChanged() {
	// 	// 			fmt.Printf("table: %s\n", result.TableMetadata().String())
	// 	// 		}
	// 	// 		// Access data
	// 	// 		fmt.Printf("value: %v\n", result.Record().Value())
	// 	// 	}
	// 	// 	// Check for an error
	// 	// 	if result.Err() != nil {
	// 	// 		fmt.Printf("query parsing error: %s\n", result.Err().Error())
	// 	// 	}
	// 	// } else {
	// 	// 	panic(err)
	// 	// }

	// Ensures background processes finishes
	influxClient.Close()
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}
	return string(result)
}
