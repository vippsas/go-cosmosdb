function triggerPostCreateInvoice() {
    let context = getContext();
    let collection = context.getCollection();
    let request = context.getRequest();
    let createdDoc = request.getBody();
    let invoiceId = createdDoc.invoiceId;
    let created = new Date().toISOString();
    let status = {
        state: "pending",
        created: created
    };
    let currentStatusDoc = {
        id: invoiceId + ".test.currentStatus",
        count: 0,
        invoiceId: invoiceId,
        modified: created,
        created: created,
        status: status
    };
    let statusesDoc = {
        id: invoiceId + ".test.statuses",
        invoiceId: invoiceId,
        statuses: [status],
        modified: created,
        created: created
    };
    let statusDoc = {
        id: invoiceId + ".test.status." + currentStatusDoc.count,
        invoiceId: invoiceId,
        status: status,
        created: created,
        isStatusDoc: true
    };
    let accepted = collection.createDocument(collection.getSelfLink(), currentStatusDoc, (err, documentCreated) => {
        if (err) {
            throw new Error("Error" + err.message);
        }
    });
    if (!accepted) {
        return;
    }
    accepted = collection.createDocument(collection.getSelfLink(), statusDoc, (err, documentCreated) => {
        if (err) {
            throw new Error("Error" + err.message);
        }
    });
    if (!accepted) {
        return;
    }
    accepted = collection.createDocument(collection.getSelfLink(), statusesDoc, (err, documentCreated) => {
        if (err) {
            throw new Error("Error" + err.message);
        }
    });
    if (!accepted) {
        return;
    }
}
