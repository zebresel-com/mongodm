[![GoDoc](https://godoc.org/github.com/zebresel-com/mongodm?status.svg)](https://godoc.org/github.com/zebresel-com/mongodm)

## What is mongodm?

The mongodm package is an object document mapper (ODM) for mongodb written in Go which uses the official mgo adapter.
You can find an **example API application** [here](https://github.com/moehlone/mongodm-example).

![Heisencat](https://octodex.github.com/images/heisencat.png)

## Features

- 1:1, 1:n struct relation mapping and embedding
- call `Save()`,`Update()`, `Delete()` and `Populate()` directly on document instances
- call `Select()`, `Sort()`, `Limit()`, `Skip()` and `Populate()` directly on querys
- validation (default and custom with regular expressions) followed by translated error list (customizable)
- population instruction possible before and after querys
- `Find()`, `FindOne()` and `FindID()`
- default handling for `ID`, `CreatedAt`, `UpdatedAt` and `Deleted` attribute
- extends `*mgo.Collection`
- default localisation (fallback if none specified)
- database authentication (user and password)
- multiple database hosts on connection

## Todos

- recursive population
- add more validation presets (like "email")
- benchmarks
- accept plain strings as objectID value
- virtuals and hooks (like in mongoose)

## Usage

### Note(!)

`Collection` naming in this package is switched to `Model`.

### Fetch (terminal)

`go get github.com/zebresel-com/mongodm`

### Import

Add `import "github.com/zebresel-com/mongodm"` in your application file.

### Define your own localisation for validation

First step is to create a language file in your application (skip if you want to use the english defaults).
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

### Create a database connection

Subsequently you have all information for mongodm usage and can now connect to a database.
Load your localisation file and parse it until you get a `map[string]string` type. Then set the database host and name. Pass the config reference to the mongodm `Connect()` method and you are done.
(You dont need to set a localisation file or credentials)

```go
	file, err := ioutil.ReadFile("locals.json")

	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	var localMap map[string]map[string]string
	json.Unmarshal(file, &localMap)

	dbConfig := &mongodm.Config{
		DatabaseHosts: []string{"127.0.0.1"},
		DatabaseName: "mongodm_sample",
		DatabaseUser: "admin",
		DatabasePassword: "admin",
		Locals:       localMap["en-US"],
	}

	connection, err := mongodm.Connect(dbConfig)

	if err != nil {
		fmt.Println("Database connection error: %v", err)
	}
```

### Create a model

```go
package models

import "github.com/zebresel-com/mongodm"

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

### Register your models (collections)

It is necessary to register your created models to the ODM to work with. Within this process
the ODM creates an internal model and type registry to work fully automatically and consistent.
Make sure you already created a connection. Registration expects a pointer to an IDocumentBase
type and the collection name where the docuements should be stored in. Register your collections only once at runtime!

For example:

```go
connection.Register(&User{}, "users")
connection.Register(&Message{}, "messages")
connection.Register(&Customer{}, "customers")
```

### Working on a model (collection)

To create actions on each collection you have to request a model instance.
Make sure that you registered your collections and schemes first, otherwise it will panic.

For example:

```go
User := connection.Model("User")

User.Find() ...
```

### Persist a new document in a collection

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

### FindOne

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

### Find

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

### FindId

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

### Populate

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
		fmt.Println("something went wrong during type assertion. wrong type?")
	}
}
```
or after your query only for single users:

```go
User := connection.Model("User")

user := &models.User{}

err := User.Find(bson.M{"firstname" : "Max"}).Exec(user)

if err != nil {
	fmt.Println(err)
}

for index, user := range users {

	if user.FirstName == "Max" {

		err := user.Populate("Messages")

		if err != nil {
		
			fmt.Println(err)
			
		} else if messages, ok := user.Messages.([]*models.Message); ok {
	
			for _, message := range messages {
	
				fmt.Println(message.Text)
			}
		} else {
			fmt.Println("something went wrong during type assertion. wrong type?")
		}
	}
}
```

Note: Only the first relation level gets populated! This process is not recursive.

### Default document validation

To validate model attributes/values you first have to define some rules.
Therefore you can add **tags**:

```go
type User struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`

	FirstName    string   `json:"firstname"  bson:"firstname" minLen:"2" maxLen:"30" required:"true"`
	LastName     string   `json:"lastname"  bson:"lastname" minLen:"2" maxLen:"30" required:"true"`
	UserName     string   `json:"username"  bson:"username" minLen:"2" maxLen:"15"`
	Email        string   `json:"email" bson:"email" validation:"email" required:"true"`
	PasswordHash string   `json:"-" bson:"passwordHash"`
	Address      *Address `json:"address" bson:"address"`
}
```

This User model defines, that the firstname for example must have a minimum length of 2 and a maximum length of 30 characters (**minLen**, **maxLen**). Each **required** attribute says, that the attribute can not be default or empty (default value is required:"false"). The **validation** tag is used for regular expression validation. Currently there is only one preset "email". A use case would be to validate the model after a request was mapped:

```go
User := self.db.Model("User")
user := &models.User{}

err, _ := User.New(user, self.Ctx.Input.RequestBody)

if err != nil {
	self.response.Error(http.StatusBadRequest, err)
	return
}

if valid, issues := user.Validate(); valid {

		err = user.Save()
		
		if err != nil {
			self.response.Error(http.StatusInternalServerError)
			return
		}
		
		// Go on..
		
	} else {
		self.response.Error(http.StatusBadRequest, issues)
		return
	}
```

This example maps a received `Ctx.Input.RequestBody` to the attribute values of a new user model. Continuing with calling `user.Validate()` we detect if the document is valid and if not what issues we have (a list of validation errors). Each `Save` call will also validate the current state. The document gets only persisted when there were no errors.

### Custom document validation

In some cases you may want to validate request parameters which do not belong to the model itself or you have to do advanced validation checks. Then you can hook up before default validation starts:

```go
func (self *User) Validate(values ...interface{}) (bool, []error) {

	var valid bool
	var validationErrors []error

	valid, validationErrors = self.DefaultValidate()

	type m map[string]string

	if len(values) > 0 {

		//expect password as first param then validate it with the next rules
		if password, ok := values[0].(string); ok {

			if len(password) < 8 {

				self.AppendError(&validationErrors, mongodm.L("validation.field_minlen", "password", 8))

			} else if len(password) > 50 {

				self.AppendError(&validationErrors, mongodm.L("validation.field_maxlen", "password", 50))
			}

		} else {

			self.AppendError(&validationErrors, mongodm.L("validation.field_required", "password"))
		}
	}

	if len(validationErrors) > 0 {
		valid = false
	}

	return valid, validationErrors
}
```
Simply add a `Validate` method in your `IDocumentBase` type model with the signature
`Validate(...interface{}) (bool, []error)`. Within this you can implement any checks that you want. You can call the `DefaultValidate` method first to run all default validations. You will get a `valid` and `validationErrors` return value.
Now you can run your custom checks and append some more errors with `AppendError(*[]error, message string)`. Also have a look at the `mongodm.L` method if you need language localisation! The next example shows how we can use our custom validate method:

```go
User := self.db.Model("User")
	user := &models.User{}

	// NOTE: we now want our request body get back as a map (requestMap)..
	err, requestMap := User.New(user, self.Ctx.Input.RequestBody)

	if err != nil {
		self.response.Error(http.StatusBadRequest, err)
		return
	}

	//NOTE: ..and validate the "password" parameter which is not part of the model/document
	if valid, issues := user.Validate(requestMap["password"]); valid {
		err = user.Save()
		
		if err != nil {
			self.response.Error(http.StatusInternalServerError)
			return
		}
	} else {
		self.response.Error(http.StatusBadRequest, issues)
		return
	}

	self.response.AddContent("user", user)
	self.response.SetStatus(http.StatusCreated)
	self.response.ServeJSON()
}
```
In this case we retrieve a `requestMap` and forward the `password` attribute to our `Validate` method (example above). 
If you want to use your own regular expression as attribute tags then use the following format: `validation:"/YOUR_REGEX/YOUR_FLAG(S)"` - for example: `validation:"/[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}/"`

## Questions?

Are there any questions or is something not clear enough? Simply open up a ticket or send me an email :)


**Also feel free to contribute! Start pull requests against the `develop` branch.**
