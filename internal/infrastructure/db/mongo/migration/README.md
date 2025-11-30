# MongoDB migrations

This folder contains migration scripts intended to be run against a MongoDB instance (using `mongosh` or the legacy `mongo` shell).

Usage examples:

- Run with `mongosh` (recommended):

```bash
mongosh "mongodb://localhost:27017/app" migrations/mongo/0002_create_products_collection.js
```

- Or, with the older `mongo` shell:

```bash
mongo "mongodb://localhost:27017/app" migrations/mongo/0002_create_products_collection.js
```

Notes:
- The script creates a `products` collection with a JSON schema validator and helpful indexes.
- Replace the connection string and DB name (`app`) as needed for your environment.
- For production, integrate these steps into your migration tooling or CI/CD pipeline.
