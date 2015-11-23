## What is mongodm?

The mongodm package is an object document mapper (ODM) for mongodb written in Go which uses the official mgo adapter.

***README file only 70% complete (!) - An example JSON API project with beego is in progress***

Maybe read API documentation instead:

[![GoDoc](https://godoc.org/github.com/zebresel-com/mongodm?status.svg)](https://godoc.org/github.com/zebresel-com/mongodm)

![Heisencat](https://octodex.github.com/images/heisencat.png)

## Features

- 1:1, 1:n struct relation mapping and embedding
- call `Save()`,`Update()`, `Delete()` and `Populate()` directly on document instances
- call `Select()`, `Sort()`, `Limit()`, `Skip()` and `Populate()` directly on querys
- validation (default and custom) followed by translated error list (customizable)
- population instruction possible before and after querys
- `Find()`, `FindOne()` and `FindID()`
- default handling for `ID`, `CreatedAt`, `UpdatedAt` and `Deleted` attribute
- extends `*mgo.Collection`

##Usage

###Note(!)

`Collection` naming in this package is switched to `Model`.

###Fetch

Call `go get github.com/zebresel-com/mongodm` in your terminal.

###Import

Add `import "github.com/zebresel-com/mongodm"` in your application file.

###Define your own localisation for validation

First step is to create a language file in your application.
This is necessary for document validation which is always processed.
The following entrys are all keys which are currently used. If one of the keys is not defined the output will be the key itself. In the next step you have to specify a translation map when creating a database connection. 

For example:

```json
{
    "en-US": {
        "validation.field_required": "Field '%s' is required.",
        "validation.field_invalid": "Field '%s' has an invalid value.",
        "validation.field_invalid_id": "Field '%s' contains an invalid object id value.",
        "validation.field_minlen": "Field '%s' must be at least %v characters long.",
        "validation.field_maxlen": "Field '%s' can be maximum %v characters long.",
        "validation.entry_exists": "%s already exists for value '%v'.",
        "validation.field_not_exclusive": "Only one of both fields can be set: '%s'' or '%s'.",
        "validation.field_required_exclusive": "Field '%s' or '%s' required."
    }
}
```

###Create a database connection

Subsequently you have all information for mongodm usage and can now connect to a database.
Load your localisation file and parse it until you get a `map[string]string` type. Then set the database host and name. Pass the config reference to the mongodm `Connect()` method and you are done.

```go
	file, err := ioutil.ReadFile("locals.json")

	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	var localMap map[string]map[string]string
	json.Unmarshal(file, &localMap)

	dbConfig := &mongodm.Config{
		DatabaseHost: "127.0.0.1",
		DatabaseName: "mongodm_sample",
		Locals:       localMap["en-US"],
	}

	connection, err := mongodm.Connect(dbConfig)

	if err != nil {
		fmt.Println("Database connection error: %v", err)
	}
```

###Create a model

```go
type User struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`

	FirstName string       `json:"firstname" bson:"firstname"`
	LastName  string       `json:"lastname"	 bson:"lastname"`
	UserName  string       `json:"username"	 bson:"username"`
	Messages  interface{}  `json:"messages"	 bson:"messages" 	model:"Message" relation:"1n" autosave:"true"`
}
```

It is important that each schema embeds the IDocumentBase type (mongodm.DocumentBase) and make sure that it is tagged as 'inline' for json and bson.
This base type also includes the default values id, createdAt, updatedAt and deleted. Those values are set automatically from the ODM.
The given example also uses a relation (User has Messages). Relations must always be from type interface{} for storing bson.ObjectId OR a completely
populated object. And of course we also need the related model for each stored message:

```go
type Message struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`

	Sender 	  string       `json:"sender" 	 bson:"sender"`
	Receiver  string       `json:"receiver"	 bson:"receiver"`
	Text  	  string       `json:"text"	 bson:"text"`
}
```
Note that when you are using relations, each model will be stored in his own collection. So the values are not embedded and instead stored as object ID
or array of object ID's.

To configure a relation the ODM understands three more tags:

	model:"Message"

		This must be the struct type you want to relate to.

		Default: none, must be set

	relation:"1n"

		It is important that you specify the relation type one-to-one or one-to-many because the ODM must decide whether it accepts an array or object.

		Possible: "1n", "11"
		Default: "11"

	autosave:"true"

		If you manipulate values of the message relation in this example and then call 'Save()' on the user instance, this flag decides if this is possible or not.
		When autosave is activated, all relations will also be saved recursively. Otherwise you have to call 'Save()' manually for each relation.

		Possible: "true", "false"
		Default: "false"

But it is not necessary to always create relations - you also can use embedded types:

```go
type Customer struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`

	FirstName string       `json:"firstname" bson:"firstname"`
	LastName  string       `json:"lastname"	 bson:"lastname"`
	Address   *Address     `json:"address"	 bson:"address"`
}

type Address struct {

	City 	string       `json:"city" 	 bson:"city"`
	Street  string       `json:"street"	 bson:"street"`
	ZipCode	int16	     `json:"zip"	 bson:"zip"`
}
```

Persisting a customer instance to the database would result in embedding a complete address object. You can embed all supported types.

Now that you got some models and a connection to the database you have to register these models for the ODM for working with them.

###Register your models (collections)

It is necessary to register your created models to the ODM to work with. Within this process
the ODM creates an internal model and type registry to work fully automatically and consistent.
Make sure you already created a connection. Registration expects a pointer to an IDocumentBase
type and the collection name where the docuements should be stored in.

For example:

```go
connection.Register(&User{}, "users")
connection.Register(&Message{}, "messages")
connection.Register(&Customer{}, "customers")
```

###Working on a model (collection)

To create actions on each collection you have to request a model instance.
Make sure that you registered your collections and schemes first, otherwise it will panic.

For example:

```go
User := connection.Model("User")

User.Find() ...
```

###Persist a new document in a collection

`Save()` persists all changes for a document. Populated relations are getting converted to object ID's / array of object ID's so you dont have to handle this by yourself.
Use this function also when the document was newly created, if it is not existent the method will call insert. During the save process createdAt and updatedAt gets also automatically persisted.

For example:

```go
User := connection.Model("User")

user := &models.User{}

User.New(user) //this sets the connection/collection for this type and is strongly necessary(!) (otherwise panic)

user.FirstName = "Max"
user.LastName = "Mustermann"

err := user.Save()
```

###FindOne

If you want to find a single document by specifing query options you have to use this method. The query param expects a map (e.g. bson.M{}) and returns a query object which has to be executed manually. Make sure that you pass an IDocumentBase type to the exec function. After this you obtain the first matching object. You also can check the error if something was found.

For example:

```go
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
```

###Find

Use`Find()` if you want to fetch a set of matching documents. Like FindOne, a map is expected as query param, but you also can call this method without any arguments. When the query is executed you have to pass a pointer to a slice of IDocumentBase types.

For example:

```go
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
```

###FindId

If you have an object ID it is possible to find the matching document with this param.

For example:

```go
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
```

###Populate

This method replaces the default object ID value with the defined relation type by specifing one or more field names. After it was succesfully populated you can access the relation field values. Note that you need type assertion for this process.

For example:

```go
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
```

Note: Only the first relation level gets populated! This process is not recursive.