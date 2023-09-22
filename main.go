package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type OilPressureSensors struct {
	SensorID int64
	Section  uint8
	Phase    string
	Endpoint string
}

type Cables struct {
	CircuitID          uint8
	CircuitName        string
	CircuitVoltage     string
	Status             string
	SourceSubstation   string
	TargetSubstation   string
	CircuitType        string
	FeederNumber       uint8
	OilPressureSensors []OilPressureSensors
}

func main() {
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

	phases := []string{"red", "yellow", "blue"}
	endpoints := []string{"source", "destination"}
	status := []string{"active", "inactive"}

	// cablePressureCollection := database.Collection("cablePressure")
	// cablePressureCollection.Drop(ctx)

	// Generate and insert random data
	for i := 1; i <= 65; i++ {
		var oilPressureSensors []OilPressureSensors
		// max number of pressure sensors per cable = max no. of Sections * number of readings per
		// section * number of phases
		// e.g. 8 * 2 * 3 = 48
		for j := 1; j <= 48; j++ {
			randomPhaseIndex := rand.Intn(len(phases))
			randomEndpointIndex := rand.Intn(len(endpoints))

			sensor := OilPressureSensors{
				SensorID: rand.Int63n(48),
				Section:  uint8(rand.Intn(8)),
				Phase:    phases[randomPhaseIndex],
				Endpoint: endpoints[randomEndpointIndex],
			}
			oilPressureSensors = append(oilPressureSensors, sensor)
		}

		randomStatusIndex := rand.Intn(len(status))

		cable := Cables{
			CircuitID:          uint8(rand.Intn(65)),
			CircuitName:        randomString(8),
			CircuitVoltage:     randomString(3),
			Status:             status[randomStatusIndex],
			SourceSubstation:   randomString(8),
			TargetSubstation:   randomString(8),
			CircuitType:        "OIL-FILLED",
			FeederNumber:       uint8(rand.Intn(5)),
			OilPressureSensors: oilPressureSensors,
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
		return
	}

	sort.SliceStable(cableResult.OilPressureSensors, func(i, j int) bool {
		return cableResult.OilPressureSensors[i].Section < cableResult.OilPressureSensors[j].Section
	})

	fmt.Println("CircuitID: ", cableResult.CircuitID)
	fmt.Println("Status: ", cableResult.Status)
	for i := range cableResult.OilPressureSensors {
		p := cableResult.OilPressureSensors[i]
		fmt.Println("Section: ", p.Section)
	}
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
