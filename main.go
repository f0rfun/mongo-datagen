package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

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
	cableCollection := database.Collection("cable")
	cableCollection.Drop(ctx)

	cablePressureCollection := database.Collection("cablePressure")
	cablePressureCollection.Drop(ctx)

	// insert one cable
	aCable := bson.M{
		"gemsID":           "4002",
		"circuitName":      "Ayer Rajah - Labrador F2",
		"sourceSubstation": "Ayer Rajah 400kV",
		"targetSubstation": "Labrador 400kV",
		"status":           "active",
		"voltageLevel":     "400",
		"feederNumber":     "2",
		"numberOfSection":  "5",
		"circuitType":      "oil-filled",
		"assetType":        "CABLE",
		"createdAt":        time.Now(),
		"createdBy":        "me",
		"updatedAt":        time.Now(),
		"updatedBy":        "me",
	}
	insertedCableResult, err := cableCollection.InsertOne(ctx, aCable)

	if err != nil {
		log.Fatalf("inserted error : %v", err)
		return
	}
	fmt.Println("======= inserted cable id ================")
	log.Printf("inserted ID is : %v", insertedCableResult.InsertedID)

	// setting historical threshold
	thresholdOld := bson.M{
		"high":      10.5,
		"low":       0.05,
		"startedAt": time.Now(),
		"endedAt":   time.Now(),
		"updatedAt": nil,
	}
	thresholdNew := bson.M{
		"high":      5.5,
		"low":       0.05,
		"startedAt": time.Now(),
		"endedAt":   time.Now(),
		"updatedAt": nil,
	}
	threshold := bson.A{}
	threshold = append(threshold, thresholdOld)
	threshold = append(threshold, thresholdNew)

	// insert one cablePressure
	aCablePressure := bson.M{
		"gemsID":    nil,
		"sensorID":  "98762",
		"assetType": "CABLE_PRESSURE",
		"threshold": threshold,

		"createdAt": time.Now(),
		"createdBy": "me",
		"updatedAt": time.Now(),
		"updatedBy": "me",
	}
	insertedCablePressureResult, err := cablePressureCollection.InsertOne(ctx, aCablePressure)

	if err != nil {
		log.Fatalf("inserted error : %v", err)
		return
	}
	fmt.Println("======= inserted cable pressure id ================")
	log.Printf("inserted ID is : %v", insertedCablePressureResult.InsertedID)

	// query all cable data
	fmt.Println("== query all cable data ==")
	cursorCable, err := cableCollection.Find(ctx, options.Find())
	if err != nil {
		log.Fatalf("find collection err : %v", err)
		return
	}
	var queryResult []bson.M
	if err := cursorCable.All(ctx, &queryResult); err != nil {
		log.Fatalf("query cable result")
		return
	}

	for _, doc := range queryResult {
		fmt.Println(doc)
	}

	// query all cable pressure data
	fmt.Println("== query all cable pressure data ==")
	cursorPressure, err := cablePressureCollection.Find(ctx, options.Find())
	if err != nil {
		log.Fatalf("find collection err : %v", err)
		return
	}
	var queryCablePressureResult []bson.M
	if err := cursorPressure.All(ctx, &queryCablePressureResult); err != nil {
		log.Fatalf("query cable pressure result")
		return
	}

	for _, doc := range queryCablePressureResult {
		fmt.Println(doc)
	}

	// // insert many data
	// fmt.Println("=========== inserted many data ===============")
	// insertedManyDocument := []interface{}{
	// 	bson.M{
	// 		"name":       "Andy",
	// 		"content":    "new test content",
	// 		"bank_money": 1500,
	// 		"create_at":  time.Now().Add(36 * time.Hour),
	// 	},
	// 	bson.M{
	// 		"name":       "Jack",
	// 		"content":    "jack content",
	// 		"bank_money": 800,
	// 		"create_at":  time.Now().Add(12 * time.Hour),
	// 	},
	// }

	// insertedManyResult, err := cableCollection.InsertMany(ctx, insertedManyDocument)
	// if err != nil {
	// 	log.Fatalf("inserted many error : %v", err)
	// 	return
	// }

	// for _, doc := range insertedManyResult.InsertedIDs {
	// 	fmt.Println(doc)
	// }

	// fmt.Println("=========== query specific data =====================")
	// // query specific data
	// filter := bson.D{
	// 	bson.E{
	// 		Key: "bank_money",
	// 		Value: bson.D{
	// 			bson.E{
	// 				Key:   "$gt",
	// 				Value: 900,
	// 			},
	// 		},
	// 	},
	// }

	// filterCursor, err := cableCollection.Find(
	// 	ctx,
	// 	filter,
	// )
	// if err != nil {
	// 	log.Fatalf("filter query data error : %v", err)
	// 	return
	// }
	// var filterResult []bson.M
	// err = filterCursor.All(ctx, &filterResult)
	// if err != nil {
	// 	log.Fatalf("filter result %v", err)
	// 	return
	// }

	// for _, filterDoc := range filterResult {
	// 	fmt.Println(filterDoc)
	// }

	// updateManyFilter := bson.D{
	// 	bson.E{
	// 		Key:   "name",
	// 		Value: "michael",
	// 	},
	// }

	// updateSet := bson.D{
	// 	bson.E{
	// 		Key: "$set",
	// 		Value: bson.D{
	// 			bson.E{
	// 				Key:   "bank_money",
	// 				Value: 2000,
	// 			},
	// 		},
	// 	},
	// }
	// // update
	// updateManyResult, err := cableCollection.UpdateMany(
	// 	ctx,
	// 	updateManyFilter,
	// 	updateSet,
	// )
	// if err != nil {
	// 	log.Fatalf("update error : %v", err)
	// 	return
	// }

	// fmt.Println("========= updated modified count ===========")
	// fmt.Println(updateManyResult.ModifiedCount)

	// // check if updated with find solution
	// checkedCursor, err := cableCollection.Find(
	// 	ctx,
	// 	bson.D{
	// 		bson.E{
	// 			Key:   "name",
	// 			Value: "michael",
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatalf("check result error : %v", err)
	// 	return
	// }
	// var checkedResult []bson.M
	// err = checkedCursor.All(ctx, &checkedResult)
	// if err != nil {
	// 	log.Fatalf("get check information error : %v", err)
	// 	return
	// }
	// fmt.Println("=========== checked updated result ==============")
	// for _, checkedDoc := range checkedResult {
	// 	fmt.Println(checkedDoc)
	// }
	// fmt.Println("===============================")
	// // delete Many

	// deleteManyResult, err := cableCollection.DeleteMany(
	// 	ctx,
	// 	bson.D{
	// 		bson.E{
	// 			Key: "bank_money",
	// 			Value: bson.D{
	// 				bson.E{
	// 					Key:   "$lt",
	// 					Value: 1000,
	// 				},
	// 			},
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatalf("delete many data error : %v", err)
	// 	return
	// }
	// fmt.Println("===== delete many data modified count =====")
	// fmt.Println(deleteManyResult.DeletedCount)

}
