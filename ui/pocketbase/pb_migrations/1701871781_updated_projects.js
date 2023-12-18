/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("7kff2zw80a7rmbu")

  // update
  collection.schema.addField(new SchemaField({
    "system": false,
    "id": "2wfeykhs",
    "name": "selectedPlan",
    "type": "relation",
    "required": false,
    "presentable": false,
    "unique": false,
    "options": {
      "collectionId": "s7nhljkrmzzu8y6",
      "cascadeDelete": false,
      "minSelect": null,
      "maxSelect": 1,
      "displayFields": null
    }
  }))

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("7kff2zw80a7rmbu")

  // update
  collection.schema.addField(new SchemaField({
    "system": false,
    "id": "2wfeykhs",
    "name": "selectedPlan",
    "type": "relation",
    "required": true,
    "presentable": false,
    "unique": false,
    "options": {
      "collectionId": "s7nhljkrmzzu8y6",
      "cascadeDelete": false,
      "minSelect": null,
      "maxSelect": 1,
      "displayFields": null
    }
  }))

  return dao.saveCollection(collection)
})
