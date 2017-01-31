package mongodm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

const (
	DBHost              string = "127.0.0.1"
	DBName              string = "mongodm_test"
	DBUser              string = "admin"
	DBPass              string = "admin"
	DBTestCollection    string = "_testCollection"
	DBTestRelCollection string = "_testRelationCollection"
)

type (
	TestModel struct {
		DocumentBase `json:",inline" bson:",inline"`

		Name          string          `json:"name" bson:"name" required:"true" minLen:"2"`
		Number        int             `json:"number" bson:"number"`
		RequiredField string          `json:"requiredField" bson:"requiredField" required:"true"`
		SomeSlice     []string        `json:"to" bson:"to"`
		Relation21    interface{}     `json:"relationTwo" bson:"relationTwo" model:"TestRelationModel" relation:"11"`
		Relation11    interface{}     `json:"relationOne" bson:"relationOne" model:"TestRelationModel" relation:"11"`
		Relation1N    interface{}     `json:"relationMany" bson:"relationMany" model:"TestRelationModel" relation:"1n"`
		TestEmbed     *TestEmbedModel `json:"testEmbed" bson:"testEmbed"`
	}

	TestRelationModel struct {
		DocumentBase `json:",inline" bson:",inline"`
		RelationName string `json:"relationName" bson:"relationName"`
	}

	TestEmbedModel struct {
		Value string `json:"value" bson:"value"`
	}
)

var dbConnection *Connection
var testRequest = []byte(`{"testmodel" : {"Name":"Max","Number":1337}}`)
var testInvalidRequest = []byte(`{"testmodel" : {"Name":"M"}}`)
var localsFile []byte

func init() {

	var err error

	localsFile, err = ioutil.ReadFile("locals.json")

	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
}

func TestConnection(t *testing.T) {

	var localMap map[string]map[string]string
	json.Unmarshal(localsFile, &localMap)

	dbConfig := &Config{
		DatabaseHosts:    []string{DBHost},
		DatabaseName:     DBName,
		DatabaseUser:     DBUser,
		DatabasePassword: DBPass,
		Locals:           localMap["en-US"],
	}

	db, err := Connect(dbConfig)

	if err != nil {

		t.Error("DB: Connection error", err)

	} else {
		dbConnection = db

		dbConnection.Register(&TestModel{}, DBTestCollection)
		dbConnection.Register(&TestRelationModel{}, DBTestRelCollection)

		Test := dbConnection.Model("testmodel")
		TestRelation := dbConnection.Model("testrelationmodel")

		//clear other entrys
		Test.RemoveAll(nil)
		TestRelation.RemoveAll(nil)
	}
}

func TestConnectionWithoutExtendedConfig(t *testing.T) {

	dbConfig := &Config{
		DatabaseHosts: []string{DBHost},
		DatabaseName:  DBName,
	}

	_, err := Connect(dbConfig)

	if err != nil {

		t.Error("DB: Connection error", err)

	}
}

func TestCreate(t *testing.T) {

	Test := dbConnection.Model("testmodel")
	TestRelation := dbConnection.Model("testrelationmodel")

	testModel := &TestModel{}
	testRelationModel := &TestRelationModel{}

	Test.New(testModel)
	TestRelation.New(testRelationModel)

	testRelationModel.RelationName = "some relation"

	errRelation := testRelationModel.Save()

	if errRelation != nil {
		t.Error("DB: creation error", errRelation)
	}

	testModel.RequiredField = "Test"
	testModel.Name = "Some test"
	testModel.Number = 1337
	testModel.Relation11 = testRelationModel
	testModel.Relation21 = nil
	testModel.Relation1N = []*TestRelationModel{testRelationModel, testRelationModel, testRelationModel, testRelationModel, testRelationModel}

	err := testModel.Save()

	if err != nil {
		t.Error("DB: creation error", err)
	}
}

func TestFindPopulate(t *testing.T) {

	Test := dbConnection.Model("testmodel")

	testModelSlice := []*TestModel{}

	err := Test.Find(bson.M{"deleted": false}).Populate("Relation11", "Relation1N").Exec(&testModelSlice)

	if len(testModelSlice) == 0 || err != nil {
		t.Error("DB: Creation error - previous test did not persist any data or find failed", err)
	}

	for _, testModel := range testModelSlice {

		if testModel.Name != "Some test" || testModel.Number != 1337 {
			t.Error("DB: test model was not saved correctly")
		}

		if _, ok := testModel.Relation11.(*TestRelationModel); !ok {
			t.Error("DB: 11 relation was not populated correctly - wrong type")
		}

		if _, ok := testModel.Relation1N.([]*TestRelationModel); !ok {
			t.Error("DB: 1N relation was not populated correctly - wrong type")
		}

		if testModel.TestEmbed != nil {
			t.Error("DB: Expected nil for embedded type")
		}

		//one loop is enough
		break
	}
}

func TestFindOneWithoutPopulate(t *testing.T) {

	Test := dbConnection.Model("testmodel")

	testModel := &TestModel{}

	err := Test.FindOne(bson.M{"deleted": false}).Exec(testModel)

	if _, ok := err.(*NotFoundError); ok {
		t.Error("DB: FindOne failed, minimum one result was expected", err)
	} else if err != nil {
		t.Error("DB: FindOne failed", err)
	}

	//nothing populated, so check for bson object id casts

	if testModel.Relation21 != nil {
		t.Error("DB: expected second relation as nil")
	}

	if _, ok := testModel.Relation11.(bson.ObjectId); !ok {
		t.Error("DB: nothing populated - cast to bson object id failed for relation 11")
	}

	if _, ok := testModel.Relation1N.([]bson.ObjectId); !ok {
		t.Error("DB: nothing populated - cast to bson object id slice failed for relation 1N")
	}
}

func TestUpdate(t *testing.T) {

	Test := dbConnection.Model("testmodel")

	testModel := &TestModel{}

	err := Test.FindOne(bson.M{"deleted": false}).Exec(testModel)

	if _, ok := err.(*NotFoundError); ok {
		t.Error("DB: FindOne failed, minimum one result was expected", err)
	} else if err != nil {
		t.Error("DB: FindOne failed", err)
	}

	testModel.Name = "Some other test"

	err = testModel.Save()

	if err != nil {
		t.Error("DB: update on model failed, save error", err)
	}

	//rollback

	testModelUpdate := &TestModel{}

	err = Test.FindOne(bson.M{"name": testModel.Name}).Exec(testModelUpdate)

	if err != nil {
		t.Error("DB: find with updated data failed", err)
	}

	testModelUpdate.Name = "Some test"

	err = testModelUpdate.Save()

	if err != nil {
		t.Error("DB: rollback on model failed, save error", err)
	}
}

func TestSetNewRelationAndPopulate(t *testing.T) {

	Test := dbConnection.Model("testmodel")
	TestRelation := dbConnection.Model("testrelationmodel")

	testModel := &TestModel{}

	err := Test.FindOne(bson.M{"deleted": false}).Exec(testModel)

	if _, ok := err.(*NotFoundError); ok {
		t.Error("DB: FindOne failed, minimum one result was expected", err)
	} else if err != nil {
		t.Error("DB: FindOne failed", err)
	}

	newTestModel := &TestRelationModel{}

	TestRelation.New(newTestModel)

	err = newTestModel.Save()

	if err != nil {
		t.Error("DB: Save for new 11 relation failed", err)
	}

	testModel.Relation11 = newTestModel.Id

	err = testModel.Save()

	if err != nil {
		t.Error("DB: Save for new 11 relation failed", err)
	}

	err = testModel.Populate("Relation11")

	if err != nil {
		t.Error("DB: Could not populate new relation", err)
	}

	if _, ok := testModel.Relation11.(*TestRelationModel); !ok {
		if err != nil {
			t.Error("DB: Population went wrong - could not cast to the new relation type")
		}
	}

	//add object id to old model for testing
	if relations, ok := testModel.Relation1N.([]bson.ObjectId); !ok {
		if err != nil {
			t.Error("DB: Cast for non population went wrong, bson object id slice expected")
		}
	} else {

		relations = append(relations, newTestModel.Id)

		testModel.Relation1N = relations

		err := testModel.Save()

		if err != nil {
			t.Error("DB: Could not save new 1n relation", err)
		}
	}
}

func TestRemove(t *testing.T) {

	Test := dbConnection.Model("testmodel")

	testModel := &TestModel{}

	err := Test.FindOne(bson.M{"deleted": false}).Exec(testModel)

	if _, ok := err.(*NotFoundError); ok {
		t.Error("DB: FindOne failed, minimum one result was expected", err)
	} else if err != nil {
		t.Error("DB: FindOne failed", err)
	}

	err = testModel.Delete()

	if err != nil {
		t.Error("DB: model could not be deleted", err)
	}

	err = Test.FindOne(bson.M{"deleted": true}).Exec(testModel)

	if _, ok := err.(*NotFoundError); ok {
		t.Error("DB: FindOne failed, expected at least one deleted model", err)
	} else if err != nil {
		t.Error("DB: FindOne failed", err)
	}
}

func TestMapping(t *testing.T) {

	Test := dbConnection.Model("testmodel")

	testModel := &TestModel{}

	err, mapped := Test.New(testModel, testRequest)

	if err != nil {
		t.Error("DB: JSON mapping failed", err)
	}

	if _, ok := mapped["testmodel"]; !ok {
		t.Error("DB: JSON mapping failed, expected also a non empty map result")
	}

	if testModel.Name != "Max" && testModel.Number != 1337 {
		t.Error("DB: JSON mapping failed, attributes where not set")
	}
}

func TestValidation(t *testing.T) {

	Test := dbConnection.Model("testmodel")

	testModel := &TestModel{}

	err, _ := Test.New(testModel, testInvalidRequest)

	if err != nil {
		t.Error("DB: JSON mapping failed", err)
	}

	if valid, issues := testModel.Validate(); valid {
		t.Error("DB: model validation failed, expected invalid request")
	} else if len(issues) != 3 {
		t.Error("DB: model validation failed, expected two issues", issues)
	}
}
