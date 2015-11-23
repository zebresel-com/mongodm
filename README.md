## What is mongodm?

The mongodm package is an object document mapper (ODM) for mongodb written in Go which uses the official mgo adapter.

***(!) README file is work in progress***

API documentation can be found here:

[![GoDoc](https://godoc.org/github.com/zebresel-com/mongodm?status.svg)](https://godoc.org/github.com/zebresel-com/mongodm)

![Heisencat](https://octodex.github.com/images/heisencat.png)

## Features

- 1:1, 1:n struct relation mapping and embedding
- call`Save()`,`Update()`, `Delete()` and `Populate()` directly on document instances
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

```golang
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

###Register your collections (models)