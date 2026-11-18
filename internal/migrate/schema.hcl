table "releases" {
  schema = schema.example
  column "github_id" {
    null = false
    type = varchar(255)
  }
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "repository_id" {
    null = false
    type = int
  }
  column "name" {
    null = false
    type = varchar(255)
  }
  column "url" {
    null = false
    type = varchar(255)
  }
  column "tag_name" {
    null = false
    type = varchar(255)
  }
  column "description" {
    null = false
    type = longtext
  }
  column "description_short" {
    null = false
    type = text
  }
  column "author" {
    null = true
    type = varchar(255)
  }
  column "is_prerelease" {
    null = false
    type = bool
  }
  column "released_at" {
    null = false
    type = datetime
  }
  column "created_at" {
    null = false
    type = datetime
  }
  column "updated_at" {
    null = false
    type = datetime
  }
  column "hash" {
    null     = false
    type     = bigint
    unsigned = true
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "releases_ibfk_1" {
    columns     = [column.repository_id]
    ref_columns = [table.repositories.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "repository_id" {
    columns = [column.repository_id]
  }
}
table "repositories" {
  schema = schema.example
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "github_id" {
    null = false
    type = varchar(255)
  }
  column "name" {
    null = false
    type = varchar(255)
  }
  column "url" {
    null = false
    type = varchar(255)
  }
  column "private" {
    null = false
    type = bool
  }
  column "created_at" {
    null = false
    type = datetime
  }
  column "updated_at" {
    null = false
    type = datetime
  }
  column "last_synced_at" {
    null = false
    type = datetime
  }
  column "image_url" {
    null = false
    type = varchar(255)
  }
  column "image_size" {
    null = false
    type = int
  }
  column "hash" {
    null     = false
    type     = bigint
    unsigned = true
  }
  primary_key {
    columns = [column.id]
  }
  index "github_id" {
    unique  = true
    columns = [column.github_id]
  }
}
table "repository_stars" {
  schema = schema.example
  column "repository_id" {
    null = false
    type = int
  }
  column "user_id" {
    null = false
    type = int
  }
  column "created_at" {
    null = false
    type = datetime
  }
  column "updated_at" {
    null = false
    type = datetime
  }
  column "type" {
    null = false
    type = tinyint
  }
  primary_key {
    columns = [column.repository_id, column.user_id, column.type]
  }
  foreign_key "repository_stars_ibfk_1" {
    columns     = [column.repository_id]
    ref_columns = [table.repositories.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "repository_stars_ibfk_2" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "user_id" {
    columns = [column.user_id]
  }
}
table "star_list_repositories" {
  schema = schema.example
  column "star_list_id" {
    null = false
    type = int
  }
  column "repository_id" {
    null = false
    type = int
  }
  column "updated_at" {
    null = false
    type = datetime
  }
  primary_key {
    columns = [column.star_list_id, column.repository_id]
  }
  foreign_key "star_list_repositories_ibfk_1" {
    columns     = [column.star_list_id]
    ref_columns = [table.star_lists.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "star_list_repositories_ibfk_2" {
    columns     = [column.repository_id]
    ref_columns = [table.repositories.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "repository_id" {
    columns = [column.repository_id]
  }
}
table "star_lists" {
  schema = schema.example
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "github_id" {
    null = false
    type = varchar(255)
  }
  column "user_id" {
    null = false
    type = int
  }
  column "name" {
    null = false
    type = varchar(255)
  }
  column "slug" {
    null = false
    type = varchar(255)
  }
  column "is_private" {
    null = false
    type = bool
  }
  column "created_at" {
    null = false
    type = datetime
  }
  column "updated_at" {
    null = false
    type = datetime
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "star_lists_ibfk_1" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "github_id" {
    unique  = true
    columns = [column.github_id]
  }
  index "user_id" {
    columns = [column.user_id]
  }
}
table "users" {
  schema = schema.example
  column "id" {
    null           = false
    type           = int
    auto_increment = true
  }
  column "username" {
    null = false
    type = varchar(255)
  }
  column "github_id" {
    null     = false
    type     = bigint
    unsigned = true
  }
  column "github_token" {
    null = false
    type = json
  }
  column "last_synced_at" {
    null = false
    type = datetime
  }
  column "public_id" {
    null = false
    type = varchar(255)
  }
  column "is_onboarded" {
    null = false
    type = bool
  }
  column "is_public" {
    null = false
    type = bool
  }
  primary_key {
    columns = [column.id]
  }
  index "github_id" {
    unique  = true
    columns = [column.github_id]
  }
  index "public_id" {
    columns = [column.public_id]
  }
}
schema "example" {
  charset = "utf8mb4"
  collate = "utf8mb4_0900_ai_ci"
}
