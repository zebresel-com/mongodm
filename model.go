package mongodm

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
The Model type stores a databse connection and single collection for a specific type (e.g. "users").
New model types can be registered with the help of the connection (see func (*Connection) Register).
Also an instance of this type embeds the default *mgo.Collection functionallity so you can call all native
mgo collection API`s, too.
*/
type Model struct {
	*mgo.Collection
	connection *Connection
}

/*
To initialize a document for a specific collection you have to call this method. Afterwards you can call all
ODM functions on the document instance.

For example:
	User := connection.Model("User")

	user := &models.User{}

	User.New(user)

	user.LastName = "Mustermann"

	user.Save() //this won`t be possible before initializing with User.New()
*/

func (self *Model) New(document IDocumentBase, content ...interface{}) (error, map[string]interface{}) {

	if document == nil {
		panic("model can not be nil")
	}

	//init collection and set pointer to its own collection (this is needed for odm operations)

	document.SetCollection(self.Collection)
	document.SetDocument(document)
	document.SetConnection(self.connection)

	if len(content) > 0 {
		return document.Update(content[0])
	}

	return nil, nil
}

/*
If you have an object ID it is possible to find the matching document with this param.

For example:
	User := connection.Model("User")

	user := &models.User{}

	err := User.FindId(bson.ObjectIdHex("55dccbf4113c615e49000001")).Select("firstname").Exec(user)

	if _, ok := err.(*mongodm.NotFoundError); ok {
		//no records were found
	} else if err != nil {
		//database error
	} else {
		fmt.Println("%v", user)
	}

*/
func (self *Model) FindId(id bson.ObjectId) *Query {

	return &Query{
		collection: self.Collection,
		connection: self.connection,
		query:      bson.M{"_id": id},
		multiple:   false,
	}
}

/*
If you want to find a single document by specifing query options you have to use this method. The query param expects a map (e.g. bson.M{}) and
returns a query object which has to be executed manually. Make sure that you pass an IDocumentBase type to the exec function.
After this you obtain the first matching object. You also can check the error if something was found.

For example:
	User := connection.Model("User")

	user := &models.User{}

	err := User.FindOne(bson.M{"firstname" : "Max", "deleted" : false}).Populate("Messages").Exec(user)

	if _, ok := err.(*mongodm.NotFoundError); ok {
		//no records were found
	} else if err != nil {
		//database error
	} else {
		fmt.Println("%v", user)
	}
*/
func (self *Model) FindOne(query ...interface{}) *Query {

	var finalQuery interface{}

	//accept zero or one query param
	if len(query) == 0 {
		finalQuery = bson.M{}
	} else if len(query) == 1 {
		finalQuery = query[0]
	} else {
		panic("DB: Find method accepts no or maximum one query param.")
	}

	return &Query{
		collection: self.Collection,
		connection: self.connection,
		query:      finalQuery,
		multiple:   false,
	}
}

/*
Use this method if you want to find a set of matching documents. Like FindOne, a map is expected as query param, but you also can call this
method without any arguments. When the query is executed you have to pass a pointer to a slice of IDocumentBase types.

For example:
	User := connection.Model("User")

	users := []*models.User{}

	err := User.Find(bson.M{"firstname" : "Max", "deleted" : false}).Populate("Messages").Exec(&users)

	if _, ok := err.(*mongodm.NotFoundError); ok { //you also can check the length of the slice
		//no records were found
	} else if err != nil {
		//database error
	} else {
		for user, _ := range users {
			fmt.Println("%v", user)
		}
	}
*/
func (self *Model) Find(query ...interface{}) *Query {

	var finalQuery interface{}

	//accept zero or one query param
	if len(query) == 0 {
		finalQuery = bson.M{}
	} else if len(query) == 1 {
		finalQuery = query[0]
	} else {
		panic("DB: Find method accepts no or maximum one query param.")
	}

	return &Query{
		query:      finalQuery,
		collection: self.Collection,
		connection: self.connection,
		multiple:   true,
	}
}
