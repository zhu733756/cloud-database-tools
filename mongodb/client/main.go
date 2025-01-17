package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// 定义命令行参数
var uri string
var createUserFlag bool
var insertFlag bool
var readFlag bool
var removeFlag bool
var numDocuments int

func init() {
	// 默认URI设置为localhost，用户可以通过命令行参数--uri来覆盖
	flag.StringVar(&uri, "uri", "mongodb://localhost:27017", "MongoDB connection URI")
	flag.BoolVar(&createUserFlag, "user", false, "Test creating a random user")
	flag.BoolVar(&insertFlag, "insert", false, "Test inserting random data")
	flag.BoolVar(&readFlag, "read", false, "Test inserting and reading random data")
	flag.BoolVar(&removeFlag, "remove", false, "Test inserting and removing random data")
	flag.IntVar(&numDocuments, "num", 1, "Number of random documents to insert")
}

func main() {
	flag.Parse() // 解析命令行参数

	// 创建MongoDB客户端
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 连接到MongoDB集群
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB cluster: %v", err)
	}
	defer client.Disconnect(ctx)

	// 确认连接成功，发送ping命令
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Failed to ping MongoDB cluster: %v", err)
	}

	fmt.Println("Connected to MongoDB cluster successfully!")

	if createUserFlag {
		createRandomUser(client)
	}

	if insertFlag {
		insertRandomDocuments(client, numDocuments)
	}

	if readFlag {
		insertAndReadRandomDocuments(client, numDocuments)
	}

	if removeFlag {
		insertAndRemoveRandomDocuments(client, numDocuments)
	}
}

func createRandomUser(client *mongo.Client) {
	collection := client.Database("admin").Collection("system.users")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rand.Seed(time.Now().UnixNano())
	username := fmt.Sprintf("user%d", rand.Int())
	password := fmt.Sprintf("password%d", rand.Int())

	user := bson.M{"user": username, "pwd": password, "roles": []bson.M{{"role": "readWrite", "db": "testdb"}}}
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("User created with username: %s and password: %s\n", username, password)
}

func insertRandomDocuments(client *mongo.Client, num int) {
	collection := client.Database("testdb").Collection("testdata")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var documents []interface{}
	for i := 0; i < num; i++ {
		documents = append(documents, bson.M{"name": fmt.Sprintf("Kimi%d", i), "age": rand.Intn(100)})
	}

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		log.Fatalf("Failed to insert documents: %v", err)
	}
	fmt.Printf("Inserted %d documents with IDs: %v\n", len(result.InsertedIDs), result.InsertedIDs)
}

func insertAndReadRandomDocuments(client *mongo.Client, num int) {
	collection := client.Database("testdb").Collection("testdata")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var documents []interface{}
	for i := 0; i < num; i++ {
		documents = append(documents, bson.M{"name": fmt.Sprintf("Kimi%d", i), "age": rand.Intn(100)})
	}

	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		log.Fatalf("Failed to insert documents: %v", err)
	}

	var results []bson.M
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to read documents: %v", err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem bson.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Documents read from database:", results)
}

func insertAndRemoveRandomDocuments(client *mongo.Client, num int) {
	collection := client.Database("testdb").Collection("testdata")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var documents []interface{}
	for i := 0; i < num; i++ {
		documents = append(documents, bson.M{"name": fmt.Sprintf("Kimi%d", i), "age": rand.Intn(100)})
	}

	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		log.Fatalf("Failed to insert documents: %v", err)
	}

	filter := bson.M{"name": bson.M{"$in": []string{"Kimi0"}}}

	deleteResult, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Fatalf("Failed to delete documents: %v", err)
	}
	fmt.Printf("Deleted %d documents\n", deleteResult.DeletedCount)
}
