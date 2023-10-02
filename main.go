package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Sensor struct {
	SensorID       int64
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
	Type string
	Data json.RawMessage
}

type PressurePoint struct {
	ID          string `json: "id"`
	Description string `json: "description"`
	// MeasuredAt   time.Time `json: "measuredAt"`
	// SentAt       time.Time `json: "sentAt"`
	Value        float64 `json:"value,string"`
	AlertLow     float64 `json: "alertLow,string"`
	CriticalLow  float64 `json: "criticalLow,string"`
	OorLow       float64 `json: "oorLow,string"`
	AlertHigh    float64 `json: "alertHigh,string"`
	CriticalHigh float64 `json: "criticalHigh,string"`
	OorHigh      float64 `json: "oorHigh,string"`
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
		points := make([]PressurePoint, 10)
		err = json.Unmarshal(data.Data, &points)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(points[0].Value)
		// temperature etc
	}

	// 	TOTAL_CIRCUITS := 65
	// 	influxConnString := "http://localhost:8086"
	// 	// influxTokenString := "Vyw31lcDtuBv66XPZ7fBL-m49ruPioGszCMZ8Yrbz2b3SsLn9DvBU7zaU9KHs4ooWdIbhUs6gWKoPL-CZPMAIg=="
	// 	influxOrgString := "sample-org"
	// 	influxBucketString := "sample-bucket"
	// 	influxTokenString := os.Getenv("INFLUXDB_TOKEN")

	// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// 	mongoClient, err := mongo.Connect(
	// 		ctx,
	// 		options.Client().ApplyURI("mongodb://jhwu:secret@localhost:27017/"),
	// 	)

	// 	defer func() {
	// 		cancel()
	// 		if err := mongoClient.Disconnect(ctx); err != nil {
	// 			log.Fatalf("mongodb disconnect error : %v", err)
	// 		}
	// 	}()

	// 	if err != nil {
	// 		log.Fatalf("connection error :%v", err)
	// 		return
	// 	}

	// 	err = mongoClient.Ping(ctx, readpref.Primary())
	// 	if err != nil {
	// 		log.Fatalf("ping mongodb error :%v", err)
	// 		return
	// 	}
	// 	fmt.Println("ping success")

	// 	// database and collection
	// 	database := mongoClient.Database("asset")
	// 	cableCollection := database.Collection("cables")
	// 	cableCollection.Drop(ctx)

	// 	status := []string{"active", "inactive"}
	// 	phaseLocations := []string{"RED_SOURCE", "RED_TARGET", "YELLOW_SOURCE", "YELLOW_TARGET", "BLUE_SOURCE", "BLUE_TARGET"}

	// 	// Generate and insert random data
	// 	for i := 1; i <= TOTAL_CIRCUITS; i++ {
	// 		randomStatusIndex := rand.Intn(len(status))

	// 		cable := Cables{
	// 			CircuitID:        int(rand.Intn(65)),
	// 			CircuitName:      randomString(8),
	// 			CircuitVoltage:   randomString(3),
	// 			Status:           status[randomStatusIndex],
	// 			SourceSubstation: randomString(8),
	// 			TargetSubstation: randomString(8),
	// 			CircuitType:      "OIL-FILLED",
	// 			FeederNumber:     int(rand.Intn(5)),
	// 			NoOfSections:     8,
	// 		}

	// 		for j := 1; j <= cable.NoOfSections; j++ {
	// 			var section Section
	// 			section.SectionNumber = j
	// 			var phase Sensor

	// 			for k := 0; k < len(phaseLocations); k++ {
	// 				phase = Sensor{
	// 					SensorID:       rand.Int63n(48),
	// 					PhaseLocations: phaseLocations[k],
	// 					Value:          float64(rand.Int63n(100)),
	// 				}

	// 				section.Sensors = append(section.Sensors, phase)
	// 			}

	// 			cable.Sections = append(cable.Sections, section)
	// 		}

	// 		_, err := cableCollection.InsertOne(ctx, cable)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	}

	// 	fmt.Println("Data inserted successfully.")
	// 	fmt.Println("=========== query specific data =====================")
	// 	filter := bson.M{"circuitid": bson.M{"$eq": 23}, "status": bson.M{"$eq": "active"}}

	// 	var cableResult Cables
	// 	// Decode unmarshals BSON back into a user-defined struct
	// 	err = cableCollection.FindOne(context.TODO(), filter).Decode(&cableResult)
	// 	if err != nil {
	// 		log.Fatalf("filter query data error : %v", err)
	// 		// return
	// 	}

	// 	//influxdb
	// 	influxClient := influxdb2.NewClient(influxConnString, influxTokenString)
	// 	writeAPI := influxClient.WriteAPIBlocking(influxOrgString, influxBucketString)
	// 	queryAPI := influxClient.QueryAPI(influxOrgString)

	// 	// arbitrary tags for sensors
	// 	var tagsMap = map[int]map[string]string{
	// 		0: {
	// 			"type":    "pressure",
	// 			"cableNo": "1",
	// 			"phase":   "red",
	// 		},
	// 		1: {
	// 			"type":    "pressure",
	// 			"cableNo": "3",
	// 			"phase":   "red",
	// 		},
	// 		2: {
	// 			"type":    "pressure",
	// 			"cableNo": "1",
	// 			"phase":   "blue",
	// 		},
	// 		3: {
	// 			"type":    "pressure",
	// 			"cableNo": "2",
	// 			"phase":   "red",
	// 		},
	// 		4: {
	// 			"type":    "pressure",
	// 			"cableNo": "5",
	// 			"phase":   "yellow",
	// 		},
	// 	}

	// 	for sensorID := 0; sensorID < 5; sensorID++ {
	// 		tags := tagsMap[sensorID]
	// 		fields := map[string]interface{}{
	// 			"value":     float64(rand.Int63n(110)),
	// 			"alertLow":  float64(rand.Int63n(100)),
	// 			"alertHigh": float64(rand.Int63n(100)),
	// 		}
	// 		point := write.NewPoint("sensor_data", tags, fields, time.Now())
	// 		time.Sleep(1 * time.Second) // separate points by 1 second

	// 		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	}

	// 	// Can adjust based on sensorID
	// 	params := tagsMap[0]

	// 	query := fmt.Sprintf(`from(bucket:"sample-bucket")
	// 		|> range(start: -1h)
	// 		|> last()
	// 		|> filter(fn: (r) => r.type == "%s")
	// 		|> filter(fn: (r) => r.cableNo == "%s")
	// 		|> filter(fn: (r) => r.phase == "%s")
	// 		|> filter(fn: (r) => r._field == "value")`,
	// 		params["type"], params["cableNo"], params["phase"])

	// 	// Get QueryTableResult
	// 	result, err := queryAPI.QueryWithParams(context.Background(), query, params)
	// 	if err == nil {
	// 		// Iterate over query response
	// 		for result.Next() {
	// 			// Notice when group key has changed
	// 			if result.TableChanged() {
	// 				fmt.Printf("table: %s\n", result.TableMetadata().String())
	// 			}
	// 			// Access data
	// 			fmt.Printf("value: %v\n", result.Record().Value())
	// 		}
	// 		// Check for an error
	// 		if result.Err() != nil {
	// 			fmt.Printf("query parsing error: %s\n", result.Err().Error())
	// 		}
	// 	} else {
	// 		panic(err)
	// 	}

	// // Ensures background processes finishes
	// influxClient.Close()
}

// func randomString(length int) string {
// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))

// 	result := make([]byte, length)
// 	for i := range result {
// 		result[i] = charset[r.Intn(len(charset))]
// 	}
// 	return string(result)
// }
