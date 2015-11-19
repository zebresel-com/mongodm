package mongodm

/*
err = User.FindId(user.Id, findUser)

if _, ok := err.(*mongodm.NotFoundError); ok {
	log.W("DB: %v", err.Error())
} else if err != nil {
	log.E("DB: %v", err)
}
*/

/*
Use this error type for checking if some records were found.

For example:

	err := User.Find(bson.M{"firstname":"Paul"}).Populate("Messages").Exec(&users)

	if _, ok := err.(*mongodm.NotFoundError); ok {
		//no records were found
	} else if err != nil {
		//database error
	}
*/

type QueryError struct {
	message string
}

type NotFoundError struct {
	*QueryError
}

type DuplicateError struct {
	*QueryError
}

type ValidationError struct {
	*QueryError
	Errors []error
}

type InvalidIdError struct {
	*QueryError
}

func (self *QueryError) Error() string {
	return self.message
}
