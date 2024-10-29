/*
Basic CRUD API calls

Steps:
1. Connect to a local DB
2. Implement CRUD for DB objects
  - Create DONE
  - Read DONE
  - Update DONE
  - Delete DONE

// WSL ip: <IP>
// db conn string format: mongodb://<username>:<password>@<ip>/<database>

Create a local instance of MongoDB and start it: sudo systemctl start mongod
I have created a test database on my WSL instance, but the process is the same.
For more information to get started on MongoDB:
https://www.mongodb.com/pt-br/docs/manual/tutorial/install-mongodb-on-ubuntu/#std-label-install-mdb-community-ubuntu
*/
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Defining the data form struct. The bson:id syntax maps the fields to mongoDB fields for marshalling/unmarshalling
type inputData struct {
	id      string  `bson:"_id,omitempty"`
	name    string  `bson:"name"`
	revenue float32 `bson:"revenue"`
}

func main() {
	// How do I connect to a database?
	conn_string := "<your_string>"

	clientOptions := options.Client().ApplyURI(conn_string)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// A simple CRUD menu
	scanner := bufio.NewScanner(os.Stdin)

	// When user selects 0, the program exits
	for scanner.Text() != "0" {
		fmt.Println("CRUD operations\n" +
			"1 - Create record to be inserted\n" +
			"2 - Read record\n" +
			"3 - Update record\n" +
			"4 - Delete Record\n" +
			"0 - Exit" +
			"\nSelect your option: ")

		scanner.Scan()
		option := scanner.Text()
		// Converting the scanner object (option var) to an integer
		optionInt, err := strconv.ParseInt(option, 10, 64)
		if err != nil {
			fmt.Println("Option must be an integer")
			return
		}

		// Cases
		switch {
		// Create
		case optionInt == 1:
			userData := getInputData()
			createRecord(userData, client)
			return
		// Read
		case optionInt == 2:
			getRecordsFromCollection(client)
			return
		// Update
		case optionInt == 3:
			updateRecord(client)
			return
		// Delete
		case optionInt == 4:
			deleteRecord(client)
			return
		// Exit
		case optionInt == 0:
			return
		}
	}

}

// generic function to get input
func getInputData() inputData {
	var data inputData
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Insert the name: ")
	if scanner.Scan() {
		data.name = scanner.Text()
	}

	fmt.Println("Insert the revenue value: ")
	if scanner.Scan() {
		revenue, err := strconv.ParseFloat(scanner.Text(), 32)
		if err != nil {
			fmt.Println("Error parsing revenue: ", err)
		}

		data.revenue = float32(revenue)

	}

	return data

}

// In the future, the param received should be the form structure
func createRecord(data inputData, client *mongo.Client) {
	record := bson.M{"name": data.name, "revenue": data.revenue}
	success, err := client.Database("test_db").Collection("sales").InsertOne(context.TODO(), record)
	if err != nil {
		fmt.Println("Could not insert record!")
	}

	// Retrieve the inserted ID
	if objectid, ok := success.InsertedID.(primitive.ObjectID); ok {
		data.id = objectid.Hex()
	}

	fmt.Printf("Created record: %s", success)
}

// read all records in collection
func getRecordsFromCollection(client *mongo.Client) {
	// Slice ptr to receive all the documents
	var documents []bson.M

	all_records_cursor, err := client.Database("test_db").Collection("sales").Find(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println("Could not read")
	}

	if err := all_records_cursor.All(context.TODO(), &documents); err != nil {
		log.Panic(err)
	}

	for _, doc := range documents {
		fmt.Println(doc["_id"], doc["name"], doc["revenue"])
	}

}

// get record by ID
// func getRecordByID(client *mongo.Client, custom_id string) bson.M {
// 	// Store the result
// 	var result bson.M

// 	objectID, err := primitive.ObjectIDFromHex(custom_id)
// 	if err != nil {
// 		log.Fatal("Invalid ID format:", err)
// 	}

// 	// Adding the ObjectID to string 671aa05bf8333e0581fe6911

// 	// Creating the filter for the id
// 	filter := bson.M{"_id": objectID}

// 	query := client.Database("test_db").Collection("sales").FindOne(context.TODO(), filter).Decode(&result)
// 	if query != nil {
// 		log.Println("Document not found: ", err)

// 	}

// 	fmt.Printf("FOUND: %v\n", result)
// 	return result
// }

// get input ID for searching record
func getRecordID() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("What is the record ID? ")
	scanner.Scan()
	custom_id := scanner.Text()
	custom_id = strings.TrimSpace(custom_id)

	return custom_id
}

// update record
func updateRecord(client *mongo.Client) {
	// First, the the ID of the record to update
	retrieved_id := getRecordID()

	// Second, convert the retrieved_id to ObjectID
	objectID, err := primitive.ObjectIDFromHex(retrieved_id)
	if err != nil {
		log.Fatal("Invalid ID format:", err)
	}

	filter := bson.M{"_id": objectID}

	// Third, get the fields to be updated. For now, we'll just call the getInputData again.
	var updated_user_data = getInputData()

	// Fourth, update the document with the new data
	update := bson.M{
		"$set": bson.M{
			"name":    updated_user_data.name,
			"revenue": updated_user_data.revenue,
		},
	}

	// Fifth, execute operation
	updateOptions := options.FindOneAndUpdate().SetReturnDocument(options.After)
	updated_record := client.Database("test_db").Collection("sales").FindOneAndUpdate(context.TODO(), filter, update, updateOptions).Decode(&updated_user_data)
	if updated_record != nil {
		log.Println("Failed to update record: ", updated_record)
		return
	}

}

// delete record
func deleteRecord(client *mongo.Client) {
	// Similar to update, search for the document id
	retrieved_id := getRecordID()

	// Convert the retrieved_id to ObjectID
	objectID, err := primitive.ObjectIDFromHex(retrieved_id)
	if err != nil {
		log.Fatal("Invalid ID format:", err)
	}

	filter := bson.M{"_id": objectID}

	// Execute delete operation
	deleted, err := client.Database("test_db").Collection("sales").DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Println("Could not delete record: ", err)
		return
	}

	fmt.Println("Successfully deleted: ", deleted)

}
