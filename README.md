# nugget

Test REST APIs from the command line.

## Features

- Define the requests in a yaml file
- Capture values from the response
- Chain requests to implement full API use cases (no programming needed!)
- Own your files: no cloud storage

## Usage

```bash
% nugget run your-requests-file.yaml
```

## Minimal example

This would be a simple yaml file with a get request:

```yaml
steps:
  - name: Get TODO list
    method: GET
    url: http://mytodo.com/api/v1/todos
```

## Add headers

You can add additional headers. nugget already takes care of the `Content-type` and `Authorization` headers (it uses Bearer Authentication).

```yaml
steps:
  - name: Get TODO list
    method: GET
    url: http://mytodo.com/api/v1/todos
    header:
      some-header: some-value
```

## Capture values

Use the capture keyword to capture a list of values from the response:

```yaml
steps:
  - name: Create TODO item
    method: POST 
    url: https://mytodos.com/api/v1/todos/create
    body: |
      {
        "name": "Go shopping",
        "due": "2024-06-09"
      }
    capture:
      todo_id: .id
      todo_name: .name
```

To capture values, use the path of the value you want to capture from the response. For example, `.address.country` if you need to capture the country in the following json:

```json
{
    "id": 1234,
    "name": "John Doe",
    "address": {
        "street": "333 Embarcadero",
        "city": "San Francisco",
        "state": "CA",
        "country": "US"
    }
}
```

## Use captured values

You can use the captured values adding the variable name in a "template" like fashion: `{{ .variable-name }}`.

The capture values can be use in the following areas:

- The body json
- The ulr
- The header, wrapping the template variable in quoutes: `"{{ .some-header }}"`

## Chain several requests (use captured values)

To chain several requests, just add more steps in the yaml file starting with the `name` of the step. Use the captured values as explained before.

```yaml
steps:
  - name: Create TODO item
    method: POST
    url: https://mytodos.com/api/v1/todos/create
    body: |
      {
        "name": "Go shopping",
        "due": "2024-06-09"
      }
    capture:
      todo_id: .id
  
  - name: Update the previous TODO
    method: PUT
    url: https://mytodos.com/api/v1/todos/{{ .todo_id }}/update
    body: |
      {
        "name": "Go grocery shopping"
      } 
```

## Reference

nugget has the following keywords:

- `steps`: entry point of the yaml file to define the list of requests
- `method`: type of requests (GET, POST, PUT, DELETE)
- `url`: the endpoint url
- `header`: list of headers
- `capture`: list of variables to capture from the response

And the following pre-defined template variables:

- `{{ .uuid }}`: generates a random UUID

## Author

Isaac Benitez  
June 2024