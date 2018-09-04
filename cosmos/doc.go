// The cosmos package implements a higher-level opinionated interface
// to Cosmos.  The goal is to encourage safe programming practices,
// not to support any operation.  It is always possible to drop down
// to the lower-level REST API wrapper in cosmosapi.
//
// WARNING
//
// The package assumes that session-level consistency model is selected
// on the account.
//
// Collection
//
// The Collection type is a simple config struct where information is
// provided once. If one wants to perform inconsistent operations one
// uses Collection directly.
//
//  collection := Collection{
//    Client: cosmosapi.New(url, config, httpClient),
//    DbName: "mydb",
//    Name: "mycollection",
//    PartitionKey: "mypartitionkey",
//  }
//  var entity MyModel
//  err = collection.StaleGet(partitionKey, id, &entity)  // possibly inconsistent read
//  err = collection.RacingPut(&entity)  // can be overwritten
//
// Collection is simply a read-config struct and therefore thread-safe.
//
// Session
//
// Use a Session to enable Cosmos' session-level consistency. The
// underlying session token changes for every write to the database,
// so it is fundamentally not-thread-safe. Additionally there is a
// non-thread-safe entity cache in use.  For instance it makes sense
// to create a new Session for each HTTP request handled. It is
// possible to connect a session to an end-user of your service by
// saving and resuming the session token.
//
// You can't actually Get or Put directly on a session; instead, you
// have to start a Transaction and pass in a closure to perform these
// operations.
//
// Reason 1) To safely do Put(), you need to do Compare-and-Swap
// (CAS). To do CAS, the operation should be written in such a way
// that it can be retried a number of times. This is best expressed as
// an idempotent closure.
//
// Reason 2) By enforcing that the Get() happens as part of the closure
// we encourage writing idempotent code; where you do not build up state
// that assumes that the function only runs once.
//
// Note: The Session itself is a struct passed by value, and WithRetries(),
// WithContext() and friends return a new struct. However they will all
// share a common underlying state constructed by collection.Session().
//
// Usage:
//
//  session := collection.Session()  // or ResumeSession()
//  err := session.Transactional(func(txn cosmos.Transaction) error {
//    var entity MyModel
//    err := txn.Get(partitionKey, id, &entity)
//    if err != nil {
//      return err
//    }
//    entity.SomeCounter = entity.SomeCounter + 1
//    txn.Put(&entity)  // only registers entity for put
//    if foo {
//      return cosmos.Rollback()  // we regret the Put(), and want to return nil without commit
//    }
//    return nil // this actually does the commit and writes entity
//  })
//
// Session cache
//
// Every CAS-write through Transaction.Put() will, if successful,
// populate the session in-memory cache. This makes sense as we are
// optimistically concurrent, it is assumed that the currently running
// request is the only one running touching the entities. Example:
//
//  err := session.Transactional(func (txn cosmos.Transaction) error {
//    var entity MyModel
//    err := txn.Get(partitionKey, id, &entity)
//    if err != nil {
//      return err
//    }
//    entity.SomeCounter = entity.SomeCounter + 1
//    txn.Put(&entity)  // ...not put in cache yet, only after successful commit
//    return nil
//  })
//  if err != nil {
//    return err
//  }
//  // Cache is now populated
//
//  < snip something else that required a break in transaction, e.g., external HTTP request >
//
//  err = session.Transactional(func (txn cosmos.Transaction) error {
//    var entity MyModel
//
//    err := txn.Get(partitionKey, id, &entity)
//    // Normally, the statement above simply fetched data from the in-memory cache, populated
//    // from the closure just above. However, if the closure needs to be re-run due to another
//    // process racing us, there will be a new network access to get the updated data.
//    <...>
//  })
//
// No cache eviction is been implemented. If one is iterating over a
// lot of entities in the same Session, one should call
// session.Drop() to release memory once one is done with a given ID.
//
package cosmos
