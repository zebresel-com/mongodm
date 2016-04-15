package mongodm

import (
	"fmt"
	"reflect"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
Query is used to configure your find requests on a specific collection / model in more detail.
Each query object method returns the same query reference to enable chains. After you have finished your configuration
run the exec function (see: func (*Query) Exec).

For Example:

	users := []*models.User{}

	User.Find(bson.M{"lastname":"Mustermann"}).Populate("Messages").Skip(10).Limit(5).Exec(&users)
*/
type Query struct {
	collection *mgo.Collection
	connection *Connection
	query      interface{}
	selector   interface{}
	populate   []string
	sort       []string
	limit      int
	skip       int
	multiple   bool
}

//See: http://godoc.org/labix.org/v2/mgo#Query.Select
func (self *Query) Select(selector interface{}) *Query {

	self.selector = selector

	return self
}

//See: http://godoc.org/labix.org/v2/mgo#Query.Sort
func (self *Query) Sort(fields ...string) *Query {

	self.sort = append(self.sort, fields...)

	return self
}

//See: http://godoc.org/labix.org/v2/mgo#Query.Limit
func (self *Query) Limit(limit int) *Query {

	self.limit = limit

	return self
}

//See: http://godoc.org/labix.org/v2/mgo#Query.Skip
func (self *Query) Skip(skip int) *Query {

	self.skip = skip

	return self
}

//see: http://godoc.org/gopkg.in/mgo.v2#Query.Count
func (self *Query) Count() (n int, err error) {

	return self.collection.Find(self.query).Count()
}

/*
This method replaces the default object ID value with the defined relation type by specifing one or more field names. After it was succesfully populated
you can access the relation field values. Note that you need type assertion for this process.

For example:
	User := connection.Model("User")

	user := &models.User{}

	err := User.Find(bson.M{"firstname" : "Max"}).Populate("Messages").Exec(user)

	if err != nil {
		fmt.Println(err)
	}

	for _, user := range users {

		if messages, ok := user.Messages.([]*models.Message); ok {

			for _, message := range messages {

				fmt.Println(message.Sender)
			}
		} else {
			fmt.Println("something went wrong during cast. wrong type?")
		}
	}

Note: Only the first relation level gets populated! This process is not recursive.
*/
func (self *Query) Populate(fields ...string) *Query {

	self.populate = append(self.populate, fields...)

	return self
}

func (self *Query) Exec(result interface{}) error {

	if result == nil {
		panic("DB: No result specified")
	}

	resultType := reflect.TypeOf(result)

	/*
	 *	Check the given result type at first to determine if its a slice or struct pointer
	 */

	//expect pointer to a slice
	if resultType.Kind() == reflect.Ptr && resultType.Elem().Kind() == reflect.Slice {

		if !self.multiple {
			panic("DB: Execution expected an IDocumentBase type!")
		}

		/*
		 *	single Query execution
		 */

		mgoQuery := self.collection.Find(self.query)

		self.extendQuery(mgoQuery)

		err := mgoQuery.All(result)

		if err == mgo.ErrNotFound {

			return &NotFoundError{&QueryError{fmt.Sprintf("No records found")}}

		} else if err != nil {

			return err

		} else {

			slice := reflect.ValueOf(result).Elem()

			for index := 0; index < slice.Len(); index++ {

				current := slice.Index(index)

				self.initWithObjectId(current)
				self.initDocument(&current, &current, self.collection, self.connection)
				err := self.runPopulation(current)

				if err != nil {

					return err
				}

			}

		}
		//expect all other types - missmatch will panic later or through mgo adapter
	} else {

		if self.multiple {
			panic("DB: Execution expected a pointer to a slice!")
		}

		/*
		 *	multiple Query execution
		 */

		mgoQuery := self.collection.Find(self.query)

		self.extendQuery(mgoQuery)

		err := mgoQuery.One(result)

		if err == mgo.ErrNotFound {

			return &NotFoundError{&QueryError{fmt.Sprintf("No record found")}}

		} else if err != nil {

			return err

		}

		value := reflect.ValueOf(result)

		self.initWithObjectId(value)
		self.initDocument(&value, &value, self.collection, self.connection)

		err = self.runPopulation(value)

		if err != nil {

			return err
		}

	}

	return nil
}

//extendQuery sets all native query options if specified
func (self *Query) extendQuery(mgoQuery *mgo.Query) {

	if self.selector != nil {
		mgoQuery.Select(self.selector)
	}

	if len(self.sort) > 0 {
		mgoQuery.Sort(self.sort...)
	}

	if self.limit != 0 {
		mgoQuery.Limit(self.limit)
	}

	if self.skip != 0 {
		mgoQuery.Skip(self.skip)
	}
}

//runPopulation populates all specified fields with defined struct types
func (self *Query) runPopulation(document reflect.Value) error {
	//iterate all specified population strings
	for _, populateFieldName := range self.populate {

		//check if the field name matches with a population
		if structField, ok := document.Elem().Type().FieldByName(populateFieldName); ok {

			modelTagValue := structField.Tag.Get("model")

			//check if the relation model tag is set
			if len(modelTagValue) == 0 {
				panic(fmt.Sprintf("DB: Related model tag was not set for field '%v' in type '%v'", populateFieldName, document.Elem().Type().Name()))
			}

			//build the relation type
			relatedModel := self.connection.Model(modelTagValue)
			relatedDocument := self.connection.document(modelTagValue)
			field := document.Elem().FieldByName(populateFieldName)

			//check if the field is existent
			if !field.IsNil() {

				/*
				 *	This part detects the relationship type (object or slice)
				 * 	and decides what has to be set. We dont need the check for "multiple"
				 * 	here because this was already done in exec.
				 */

				switch fieldType := field.Interface().(type) {

				//one-to-one
				case bson.ObjectId:

					//find the matching document in the related collection
					relatedId := fieldType
					relationError := relatedModel.FindId(relatedId).Exec(relatedDocument)

					if relationError == mgo.ErrNotFound && relationError != nil {

						//dont set/init anything here, because nil is the correct behaviour
						return relationError

					} else {

						//populate the field
						value := reflect.ValueOf(relatedDocument)

						self.initWithObjectId(value)
						self.initDocument(&value, &value, relatedModel.Collection, relatedModel.connection)

						field.Set(value)
					}

				//one-to-many
				case []interface{}:

					//cast the object id slice to []bson.ObjectId
					idSlice := reflect.ValueOf(fieldType)
					idSliceInterface, _ := idSlice.Interface().([]interface{})

					//create a result slice with length of id slice (pointer is important for query execution!)
					resultSlice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(relatedDocument)), idSlice.Len(), idSlice.Len())
					resultSlicePtr := reflect.New(resultSlice.Type())

					//find relation objects by searching for ids which match with entrys from id slice
					relationError := relatedModel.Find(bson.M{"_id": bson.M{"$in": &idSliceInterface}}).Exec(resultSlicePtr.Interface())

					if relationError == mgo.ErrNotFound || resultSlice.Len() == 0 {

						//in this case it is strictly necessary to init an empty slice of the document type (nil wouldnt be correct)
						field.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(relatedDocument)), 0, 0))

					} else if relationError != nil {
						return relationError

					} else {

						field.Set(resultSlicePtr.Elem())

						for index := 0; index < resultSlicePtr.Elem().Len(); index++ {

							populatedChild := resultSlicePtr.Elem().Index(index)

							self.initWithObjectId(populatedChild)
							self.initDocument(&populatedChild, &populatedChild, relatedModel.Collection, relatedModel.connection)
						}
					}

				default:

					panic("DB: unknown type stored as relation - bson.ObjectId or []bson.ObjectId expected")
				}
			}

		} else {
			panic(fmt.Sprintf("DB: Can not populate field '%v' for type '%v'. Field not found.", populateFieldName, document.Elem().Type().Name()))
		}
	}

	return nil
}

/*
 * bson object ids have to be copied again, because after passing
 * the result reference to the mgo execute function the type
 * is overwritten as interface{} again. So initWithObjectId initializes
 * bson.ObjectId and []bson.ObjectId types.
 */
func (self *Query) initWithObjectId(document reflect.Value) {

	//If there is nothing to populate, init the fields with object id types
	if len(self.populate) == 0 {

		structElement := document.Elem()
		fieldType := structElement.Type()

		//Iterate over all struct fields
		for fieldIndex := 0; fieldIndex < structElement.NumField(); fieldIndex++ {

			relationTag := fieldType.Field(fieldIndex).Tag.Get("relation")
			field := structElement.Field(fieldIndex)

			if len(relationTag) > 0 {

				if relationTag == REL_1N {

					if field.IsNil() {

						//field.Set(reflect.ValueOf(make([]bson.ObjectId, 0, 0)))

					} else {

						slice := field.Elem()
						idSlice := make([]bson.ObjectId, slice.Len(), slice.Len())

						for index := 0; index < slice.Len(); index++ {
							idSlice[index] = slice.Index(index).Elem().Interface().(bson.ObjectId)
						}

						field.Set(reflect.ValueOf(idSlice))

					}

				} else {

					if !field.IsNil() {

						objectId := field.Elem().Interface().(bson.ObjectId)
						field.Set(reflect.ValueOf(objectId))
					}

					//field.Set(reflect.Zero(reflect.TypeOf(bson.ObjectId(""))))
				}
			}
		}
	}
}

//like Model.New(), only directly for reflect types
func (self *Query) initDocument(model *reflect.Value, document *reflect.Value, collection *mgo.Collection, connection *Connection) {

	documentMethod := model.MethodByName("SetDocument")
	collectionMethod := model.MethodByName("SetCollection")
	connectionMethod := model.MethodByName("SetConnection")

	if !documentMethod.IsValid() || !collectionMethod.IsValid() || !connectionMethod.IsValid() {
		panic("Given models were not correctly initialized with 'DocumentBase' interface type")
	}

	documentInput := []reflect.Value{*document}
	collectionInput := []reflect.Value{reflect.ValueOf(collection)}
	connectionInput := []reflect.Value{reflect.ValueOf(connection)}

	documentMethod.Call(documentInput)
	collectionMethod.Call(collectionInput)
	connectionMethod.Call(connectionInput)
}
