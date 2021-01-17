# A simple doc generator

## Required

* github.com/karta0807913/go_server_utils/tools/generate_router

## usage

run `doc_generate` in generate_router generated document folder then build it.
Generated binary can be used for generate document.

target struct

```
type Borrower struct {
    // user id
	ID uint `gorm:"primaryKey" json:"id"`
    // user name
	Name string `json:"name" gorm:"not null;index"`
    // user phone
	Phone string `json:"phone" gorm:"not null;index"`
}
```

source file `doc/README.md`

```
{{ define "ParamsTable" }}
| parameters | type | required | note |
| ---------- | ---- | -------- | ---- |{{ range . }}
| {{ .Alias }} | {{ .Type }} | {{ if .Required }} Y {{ else }} N {{ end }} | {{ .Comment }} |{{ end }}
{{ end }}

### GET `/api/borrower`

{{ template "ParamsTable" .FindBorrower.Fields }}
| limit      | number | N        | limit       |
| offset     | number | N        | offset      |

### POST `/api/borrower`

{{ template "ParamsTable" .CreateBorrower.Fields }}
```

run command `./doc/doc ./doc/README.md README.md`

output file

```
### GET `/api/borrower`

| parameters | type   | required | note        |
| ---------- | ------ | -------- | ----------- |
| name       | string | N        | user name   |
| phone      | string | N        | user phone  |
| limit      | number | N        | limit       |
| offset     | number | N        | offset      |

### POST `/api/borrower`

| parameters | type   | required | note        |
| ---------- | ------ | -------- | ----------- |
| name       | string | Y        | user name   |
| phone      | string | N        | user phone  |

```