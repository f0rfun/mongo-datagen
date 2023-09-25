package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Sensor struct {
	SensorID int64
	Value    float64
}

type Section struct {
	SectionNumber int
	RedSource     Sensor
	RedTarget     Sensor
	YellowSource  Sensor
	YellowTarget  Sensor
	BlueSource    Sensor
	BlueTarget    Sensor
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

		var section Section
		for j := 1; j <= cable.NoOfSections; j++ {
			section.SectionNumber = j
			section.RedSource = Sensor{
				SensorID: rand.Int63n(48),
				Value:    float64(rand.Int63n(48)),
			}
			section.RedTarget = Sensor{
				SensorID: rand.Int63n(48),
				Value:    float64(rand.Int63n(48)),
			}
			section.YellowSource = Sensor{
				SensorID: rand.Int63n(48),
				Value:    float64(rand.Int63n(48)),
			}
			section.YellowTarget = Sensor{
				SensorID: rand.Int63n(48),
				Value:    float64(rand.Int63n(48)),
			}
			section.BlueSource = Sensor{
				SensorID: rand.Int63n(48),
				Value:    float64(rand.Int63n(48)),
			}
			section.BlueTarget = Sensor{
				SensorID: rand.Int63n(48),
				Value:    float64(rand.Int63n(48)),
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
		return
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
