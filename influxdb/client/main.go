package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

var (
	host        string
	db          string
	measurement string
	user        string
	password    string
	count       int
	tags        int
	op          string
	tagValue    string
	query       string
)

func main() {
	flag.StringVar(&host, "host", "http://localhost:8086", "InfluxDB URI")
	flag.StringVar(&db, "db", "testdb", "InfluxDB db")
	flag.StringVar(&measurement, "measurement", "test_measurement", "InfluxDB measurement")
	flag.StringVar(&user, "user", "admint", "Username")
	flag.StringVar(&password, "password", "autotest@123", "password")
	flag.IntVar(&count, "count", 10000, "InfluxDB data count")
	flag.IntVar(&tags, "tags", 2, "Number of distinct tag sets (default 2)")
	flag.StringVar(&tagValue, "tagValue", "tag1", "tagValue")
	flag.StringVar(&op, "op", "insert", "insert/summary")
	flag.StringVar(&query, "query", "show databases", "query")

	flag.Parse()

	var c client.HTTPClient
	var err error

	c, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     host,
		Username: user,
		Password: password,
	})
	if err != nil {
		log.Fatalf("new client failed: %v", err)
	}
	defer c.Close()

	if _, _, err := c.Ping(time.Second); err != nil {
		log.Fatalf("ping failed: %v", err)
	}

	switch op {
	case "insert", "i":
		printInfo()
		err = insertData(c, db, count, tags)
	case "summary", "s":
		printInfo()
		err = summaryData(c, db)
	case "query", "q":
		err = queryData(c, db)
	default:
		err = errors.New("invalid action. Use 'insert' or 'summary'")
	}

	if err != nil {
		log.Fatalf(err.Error())
	}
}

func printInfo() {
	fmt.Printf("op:\t%s\n", op)
	fmt.Printf("db:\t%s\n", db)
	fmt.Printf("measurement:\t%s\n", measurement)
	fmt.Printf("host:\t%s\n", host)
	fmt.Printf("user:\t%s\n", user)
	fmt.Printf("password:\t%s\n", password)
	fmt.Printf("tags:\t%s\n", tagValue)
}

func insertData(c client.HTTPClient, db string, count, tags int) error {
	// 检查数据库是否存在，如果不存在则创建
	_, err := c.Query(client.NewQuery(`CREATE DATABASE IF NOT EXISTS "`+db+`"`, "", ""))
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}
	fmt.Println("Database created or already exists.")

	// 计算每个tag set的数据量
	dataPerTagSet := count / tags

	var wg sync.WaitGroup
	batchSize := 1000 // 批量大小

	for tagIndex := 0; tagIndex < tags; tagIndex++ {
		wg.Add(1)
		go func(tagIndex int) {
			defer wg.Done()
			tagsMap := map[string]string{"tags": fmt.Sprintf("tag%d", (tagIndex + 1))}

			bp, err := client.NewBatchPoints(client.BatchPointsConfig{
				Database:  db,
				Precision: "s",
			})
			if err != nil {
				log.Fatalf("NewBatchPoints error: %v", err)
			}

			for i := 0; i < dataPerTagSet; i++ {
				fields := map[string]interface{}{
					"temperature": 23.5,
					"id":          i,
					"value":       rand.Float64(),
					"loc":         rand.Float64(),
				}
				pt, err := client.NewPoint(measurement, tagsMap, fields, time.Now())
				if err != nil {
					log.Printf("NewPoint error: %v", err)
					return
				}

				bp.AddPoint(pt)
				if len(bp.Points()) >= batchSize {
					if err := c.Write(bp); err != nil {
						log.Printf("Write error: %v", err)
						return
					}
					bp, err = client.NewBatchPoints(client.BatchPointsConfig{
						Database:  db,
						Precision: "s",
					})
					if err != nil {
						log.Printf("NewBatchPoints error: %v", err)
						return
					}
				}
			}

			// 写入剩余的点
			if len(bp.Points()) > 0 {
				if err := c.Write(bp); err != nil {
					log.Printf("Write error: %v", err)
					return
				}
			}

			fmt.Printf("Data part-%d insertion complete\n", tagIndex)
		}(tagIndex)
	}

	wg.Wait()
	fmt.Println("Data insertion complete.")
	return nil
}

func summaryData(c client.HTTPClient, db string) error {
	result, err := c.Query(client.NewQuery(`SHOW MEASUREMENTS ON "`+db+`"`, db, ""))
	if err != nil {
		return fmt.Errorf("failed to query measurements: %v", err)
	}
	fmt.Println("Measurements:")
	for i := range result.Results {
		fmt.Println(result.Results[i])
	}

	result, err = c.Query(client.NewQuery(`SHOW TAG KEYS ON "`+db+`"`, db, ""))
	if err != nil {
		log.Fatalf("failed to query tag keys: %v", err)
	}
	fmt.Println("Tag Keys:")
	for i := range result.Results {
		if i > 10 {
			break
		}
		fmt.Println(result.Results[i])
	}

	result, err = c.Query(client.NewQuery(`SELECT COUNT(*) FROM "`+measurement+`"`, db, ""))
	if err != nil {
		return fmt.Errorf("failed to query count: %v", err)
	}

	fmt.Println("COUNT without tags")
	if rets := result.Results; len(rets) > 0 {
		if series := rets[0].Series; len(series) > 0 {
			if values := series[0].Values; len(values) > 0 {
				if value := values[0]; len(value) > 1 {
					fmt.Printf("COUNT: %v\n", value[1])
				}
			}
		}
	}

	result, err = c.Query(client.NewQuery(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE tags='%s'", measurement, tagValue), db, ""))
	if err != nil {
		return fmt.Errorf("failed to query count: %v", err)
	}

	fmt.Println("COUNT with tags=" + tagValue)
	if rets := result.Results; len(rets) > 0 {
		if series := rets[0].Series; len(series) > 0 {
			if values := series[0].Values; len(values) > 0 {
				if value := values[0]; len(value) > 1 {
					fmt.Printf("COUNT: %v\n", value[1])
				}
			}
		}
	}

	return nil
}

func queryData(c client.HTTPClient, db string) error {
	_, err := c.Query(client.NewQuery(`CREATE DATABASE IF NOT EXISTS "`+db+`"`, "", ""))
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	results, err := c.Query(client.NewQuery(query, db, ""))
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	b, err := json.Marshal(results)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}
