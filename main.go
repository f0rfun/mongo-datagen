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

type Sensor struct {
	SensorID       int
	Value          float64
	PhaseLocations string
}

type Section struct {
	SectionNumber int
	Sensors       []Sensor
}

type Cables struct {
	CircuitID        int
	CircuitName      string
	CircuitVoltage   string
	Status           string
	SourceSubstation string
	TargetSubstation string
	CircuitType      string
	FeederNumber     int
	NoOfSections     int
	Sections         []Section
}

type Data struct {
	Type    string
	RawData json.RawMessage `json:"data"`
}

type PressurePoint struct {
	RawSensorID  int     `json:"id"`
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
	// influxTokenString := "Vyw31lcDtuBv66XPZ7fBL-m49ruPioGszCMZ8Yrbz2b3SsLn9DvBU7zaU9KHs4ooWdIbhUs6gWKoPL-CZPMAIg=="
	influxOrgString := "sample-org"
	influxBucketString := "sample-bucket"
	influxTokenString := os.Getenv("INFLUXDB_TOKEN")

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

	status := []string{"active", "inactive"}
	phaseLocations := []string{"RED_SOURCE", "RED_TARGET", "YELLOW_SOURCE", "YELLOW_TARGET", "BLUE_SOURCE", "BLUE_TARGET"}
	var currsensorid int

	// Generate and insert random data
	for i := 1; i <= TOTAL_CIRCUITS; i++ {
		randomStatusIndex := rand.Intn(len(status))

		cable := Cables{
			CircuitID:        int(rand.Intn(65)),
			CircuitName:      randomString(8),
			CircuitVoltage:   randomString(3),
			Status:           status[randomStatusIndex],
			SourceSubstation: randomString(8),
			TargetSubstation: randomString(8),
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
					SensorID:       currsensorid,
					PhaseLocations: phaseLocations[k],
					Value:          float64(rand.Int63n(100)),
				}

				section.Sensors = append(section.Sensors, phase)
			}

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

	datajson, err := os.ReadFile("testdata/cop_format.json")
	if err != nil {
		log.Fatal(err)
	}

	var data Data
	err = json.Unmarshal(datajson, &data)
	if err != nil {
		log.Fatal(err)
	}

	SensorDataMap := make(map[int]PressurePoint)
	switch data.Type {
	case "pressure":
		var rawDataArray []PressurePoint
		err = json.Unmarshal(data.RawData, &rawDataArray)
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
						"cableID":        fmt.Sprint(cable.CircuitID),
						"section":        fmt.Sprint(section.SectionNumber),
						"phaseLocations": sensor.PhaseLocations,
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

	// 	// query := fmt.Sprintf(`from(bucket:"sample-bucket")
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
