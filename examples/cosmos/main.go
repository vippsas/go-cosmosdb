package main

import (
	"time"

	"github.com/alecthomas/repr"
	"github.com/vippsas/go-cosmosdb/cosmos"
	"github.com/vippsas/go-cosmosdb/cosmostest"
	"log"
	"os"
)

type MyModel struct {
	cosmos.BaseModel

	Model  string `json:"model" cosmosmodel:"MyModel/1"`
	UserId string `json:"userId"`
	X      int    `json:"x"`
}

func (e *MyModel) PrePut(txn *cosmos.Transaction) error {
	return nil
}

func (e *MyModel) PostGet(txn *cosmos.Transaction) error {
	return nil
}

type MyModelV2 struct {
	cosmos.BaseModel

	Model     string `json:"model" cosmosmodel:"MyModel/2"`
	UserId    string `json:"userId"`
	X         int    `json:"x"`
	TwoTimesX int    `json:"xTimes2"`
}

func (e *MyModelV2) PrePut(txn *cosmos.Transaction) error {
	return nil
}

func (e *MyModelV2) PostGet(txn *cosmos.Transaction) error {
	return nil
}

func MyModelToMyModelV2(mi1, mi2 interface{}) error {
	m1 := mi1.(MyModel)
	m2 := mi2.(*MyModelV2)
	repr.Println("conversion", m1, m2)

	return nil
}

var _ = cosmos.AddMigration(
	&MyModel{},
	&MyModelV2{},
	MyModelToMyModelV2)

type Config struct {
	Section struct {
		MasterKey string `yaml:"MasterKey"`
		Uri       string `yaml:"Uri"`
	} `yaml:"lib_cosmos_testcmd"`
}

func requireNil(err error) {
	if err != nil {
		panic(err)
	}
	return
}

func main() {

	c := cosmostest.Setup(log.New(os.Stderr, "", log.LstdFlags), "mycollection", "userId")
	defer cosmostest.Teardown(c)

	var entity MyModel
	requireNil(c.StaleGet("alice", "id2", &entity))
	repr.Println(entity)
	entity.X = entity.X + 1
	requireNil(c.RacingPut(&entity))

	session := c.Session()
	requireNil(session.Transaction(func(txn *cosmos.Transaction) error {
		var entity MyModel
		requireNil(txn.Get("alice", "id2", &entity))
		repr.Println("GET1", entity)
		entity.X = entity.X + 1
		txn.Put(&entity)
		return nil
	}))

	// Some external request...
	time.Sleep(time.Second)

	requireNil(session.Transaction(func(txn *cosmos.Transaction) error {
		var entity MyModel
		requireNil(txn.Get("alice", "id2", &entity))
		repr.Println("GET2", entity)
		entity.X = entity.X + 1
		txn.Put(&entity)
		return nil
	}))

	return
}
