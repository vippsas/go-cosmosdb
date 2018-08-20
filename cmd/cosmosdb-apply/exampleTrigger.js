function trigger() {
    let context = getContext();
    let collection = context.getCollection();
    let request = context.getRequest();
    let createdDoc = request.getBody();

    let accepted = collection.createDocument(collection.getSelfLink(), currentStatusDoc, (err, documentCreated) => {
        if (err) {
            throw err
        }
    });
}
