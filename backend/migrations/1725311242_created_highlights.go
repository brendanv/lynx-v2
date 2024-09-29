package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `{
			"id": "uxyi2qblr5y376o",
			"created": "2024-09-02 21:07:21.730Z",
			"updated": "2024-09-02 21:07:21.730Z",
			"name": "highlights",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "ojchjxxl",
					"name": "user",
					"type": "relation",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": true,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "7kxe8qib",
					"name": "link",
					"type": "relation",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "0mucz6opmdvkaqc",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "xzh2pcyt",
					"name": "start_index",
					"type": "number",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": 0,
						"max": null,
						"noDecimal": true
					}
				},
				{
					"system": false,
					"id": "88mhugxs",
					"name": "end_index",
					"type": "number",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"noDecimal": false
					}
				},
				{
					"system": false,
					"id": "sieputqe",
					"name": "highlighted_text",
					"type": "editor",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": {
						"convertUrls": false
					}
				},
				{
					"system": false,
					"id": "4053p0i6",
					"name": "tags",
					"type": "relation",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "u3528zyzzxxe55f",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": null,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "tasqrpf2",
					"name": "link_backup_url",
					"type": "url",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"exceptDomains": [],
						"onlyDomains": []
					}
				},
				{
					"system": false,
					"id": "b1wztgue",
					"name": "link_backup_hostname",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "wqti5yne",
					"name": "link_backup_title",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_6aZYAfG` + "`" + ` ON ` + "`" + `highlights` + "`" + ` (\n  ` + "`" + `link` + "`" + `,\n  ` + "`" + `start_index` + "`" + `\n)",
				"CREATE INDEX ` + "`" + `idx_mimat6g` + "`" + ` ON ` + "`" + `highlights` + "`" + ` (\n  ` + "`" + `user` + "`" + `,\n  ` + "`" + `created` + "`" + `\n)"
			],
			"listRule": "user = @request.auth.id",
			"viewRule": "user = @request.auth.id",
			"createRule": "user = @request.auth.id",
			"updateRule": "user = @request.auth.id",
			"deleteRule": "user = @request.auth.id",
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("uxyi2qblr5y376o")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
