package main

import (
	"context"
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

func main() {
	TOTAL_CIRCUITS := 65
	influxConnString := os.Getenv("INFLUXDB_URI")
	influxTokenString := os.Getenv("INFLUXDB_TOKEN")
	influxOrgString := os.Getenv("INFLUX_ORG")
	influxBucketString := os.Getenv("INFLUX_BUCKET")

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
				phase = Sensor{
					SensorID:       rand.Int63n(48),
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
	fmt.Println("=========== query specific data =====================")
	filter := bson.M{"circuitid": bson.M{"$eq": 23}, "status": bson.M{"$eq": "active"}}

	var cableResult Cables
	// Decode unmarshals BSON back into a user-defined struct
	err = cableCollection.FindOne(context.TODO(), filter).Decode(&cableResult)
	if err != nil {
		log.Fatalf("filter query data error : %v", err)
		// return
	}

	//influxdb
	influxClient := influxdb2.NewClient(influxConnString, influxTokenString)
	writeAPI := influxClient.WriteAPIBlocking(influxOrgString, influxBucketString)
	queryAPI := influxClient.QueryAPI(influxOrgString)

	// Create point using full params constructor
	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45},
		time.Now())
	// Write point immediately
	writeAPI.WritePoint(context.Background(), p)

	// Get QueryTableResult
	result, err := queryAPI.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// Access data
			fmt.Printf("value: %v\n", result.Record().Value())
		}
		// Check for an error
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}

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

func generateSensorData() map[string]interface{} {
	return map[string]interface{}{
		"p1_red_source": rand.Float64() * 100,
		"p2_red_source": rand.Float64() * 100,
	}
}
