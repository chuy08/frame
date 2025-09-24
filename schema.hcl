table "users" {
  schema = schema.public
  column "id" {
    type = uuid
    primary_key = true
  }
  column "first_name" {
    null = true
    type = varchar(100)
  }
  column "last_name" {
    null = true
    type = varchar(100)
  }
  column "email" {
    null = true
    type = varchar(100)
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}
schema "public" {
}

table "address" {
  schema = schema.public
  column "id" {
    type = uuid
    primary_key = true
  }
  column "user_id" {
    type = uuid
    primary_key = false
  }
  column "name" {
    null = true
    type = text
  }
  column "street" {
    null = true
    type = text
  }
  column "suite" {
    null = true
    type = text
  }
  column "city" {
    null = true
    type = text
  }
  column "state" {
    null = true
    type = text
  }
  column "zip" {
    null = true
    type = text
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "user_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
  }
}

table "phone" {
  schema = schema.public
  column "id" {
    type = uuid
    primary_key = true
  }
  column "user_id" {
    type = uuid
    primary_key = false
  }
  column "name" {
    null = false
    type = text
  }
  column "number" {
    null = false
    type = text
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "user_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
  }
}

table "exercise_names" {
  schema = schema.public
  column "id" {
    type = uuid
    primary_key = true
  }
  column "name" {
    null = false
    type = text
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}
