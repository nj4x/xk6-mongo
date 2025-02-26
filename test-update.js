import mongo from 'k6/x/mongo';
import files from 'k6/x/files';
import {check} from 'k6';

// const client = mongo.newClient('=mongodb://localhost:27017');
const client = mongo.newClient(`${__ENV.MONGO_URL}`, true /* unacknowledgedWriteConcern */);
const db = "testdb";
const col = "testcollection";
const maxNumberOfDocuments = 5
const filepath = 'sample-output.txt';

export function setup() {
    console.log("Setup")
    let totalDocs = client.count(db, col, {});
    console.log("totalDocs: ", totalDocs)
    let random = Math.floor(Math.random() * totalDocs);
    console.log("random: ", random)
    let randomDocs = client.find(db, col, {}, maxNumberOfDocuments, random, {"_id": 1, "updateTime": 1})
    console.log("randomDocs: ", randomDocs)
    files.delete(`nonexistent${filepath}`)
    files.delete(filepath)
    return {"objects": randomDocs}
}

export default (data) => {
    for (let obj of data.objects) {

        let updatedCount = client.updateOne(db, col, {_id: mongo.hexToObjectID(obj._id)},
            {updateTime: new Date(new Date(obj.updateTime).getTime() + 1)})
        //console.log(`[1] ${obj._id} Updated ${updatedCount} records`);

        let updatedCount2 = client.updateOne(db, col, {_id: mongo.hexToObjectID(obj._id)},
            {"updateTime": {"$add": ["$updateTime", 1]} })
        //console.log(`[2] ${obj._id} Updated ${updatedCount2} records`);

        files.append(filepath, `${obj._id}\n`);
    }
}

export function teardown(data) {
    let file = files.read(filepath);
    let ids = file.content.split("\n");
    console.log(`teardown: ids: ${ids.length}`)

    for (let obj of data.objects) {
        // func (c *Client) FindOne(database string, collection string, filter map[string]any, skip int64) error {
        let rec = client.findOne(db, col, {_id: mongo.hexToObjectID(obj._id)})
        console.log(`teardown: ${obj._id}: rec: ${rec} rec.updateTime: ${rec.updateTime}`)

        check(rec, {
            'verify timestamp': (r) => {
                const updateTimeExpected = new Date(obj.updateTime).getTime() + 2;
                const updateTimeActual = new Date(r.updateTime).getTime();

                console.log(`teardown: ${obj._id}: expected: ${updateTimeExpected}, actual: ${r.updateTime} equal?: ${updateTimeExpected === updateTimeActual}`)

                return updateTimeExpected === updateTimeActual;
            },
        });
    }
}