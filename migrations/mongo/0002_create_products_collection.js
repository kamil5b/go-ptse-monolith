// MongoDB migration: create `products` collection with JSON schema validator and indexes
// Run with mongosh or mongo shell. Example:
// mongosh "mongodb://localhost:27017/app" migrations/mongo/0002_create_products_collection.js

// switch to DB (replace 'app' with your DB name if different)
db = db.getSiblingDB(typeof db === 'object' && db._name ? db._name : 'app');

// Create collection with simple JSON schema validator to catch common mistakes
const validator = {
  $jsonSchema: {
    bsonType: 'object',
    required: ['id', 'name', 'created_at'],
    properties: {
      id: { bsonType: 'string', description: 'UUID string' },
      name: { bsonType: 'string' },
      description: { bsonType: ['string', 'null'] },
      created_at: { bsonType: 'date' },
      created_by: { bsonType: ['string', 'null'] },
      updated_at: { bsonType: ['date', 'null'] },
      updated_by: { bsonType: ['string', 'null'] },
      deleted_at: { bsonType: ['date', 'null'] },
      deleted_by: { bsonType: ['string', 'null'] }
    }
  }
};

const collName = 'products';
if (!db.getCollectionNames().includes(collName)) {
  print(`Creating collection ${collName} with validator`);
  db.createCollection(collName, { validator });
} else {
  print(`${collName} already exists — updating validator`);
  db.runCommand({ collMod: collName, validator });
}

// Create helpful indexes
print('Creating indexes on products');
db[collName].createIndex({ id: 1 }, { unique: true });
db[collName].createIndex({ name: 1 });
db[collName].createIndex({ created_at: -1 });
db[collName].createIndex({ deleted_at: 1 });

print('MongoDB migration completed.');
