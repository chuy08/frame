env "local" {
    schema {
        src = "file://schema_pg.hcl"
    }

    url = "postgres://postgres:postgres@127.0.0.1:15432/framework?sslmode=require"

    dev = "docker://postgres/17/dev?search_path=public"

    format {
        migrate {
            diff = "{{ sql . \" \" }}"
        }
    }
}