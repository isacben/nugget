# nugget

Run API requests from the command line.

## Features

- Define the requests in a file
- Capture values from the response
- Chain requests to implement full API use cases
- Own your files, no cloud storage

## Usage

```bash
% nugget requests.n
```

Use the `--raw` flag to print the raw response.

```bash
% nugget requests.n --raw
```

Pass the `-H` flag to include the response headers in the output.

```bash
% nugget requests.n -H
```

## Minimal example

This would be a simple yaml file with a get request:

```
# Get TODO list
GET http://mytodo.com/api/v1/todos
```

## Add headers

You can add additional headers. nugget already takes care of the `Content-type` and `Authorization` headers (it uses Bearer Authentication).

```
GET http://mytodo.com/api/v1/todos
header some-header some-value
```

## Add assertions

Use the `http` keyword to assert the status code of the response.

```
GET http://mytodo.com/api/v1/todos
HTTP 200
```

## Capture values

Use the capture keyword to capture a list of values from the response:

```
# Create TODO item
POST https://mytodos.com/api/v1/todos/create
{
  "name": "Go shopping",
  "due": "2024-06-09"
}
SAVE todo_id .id
SAVE todo_name .name
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

The captured values can be use in the following areas:

- The body json
- The url
- The header, wrapping the template variable in quoutes: `"{{ .some-header }}"`

## Chain several requests (use captured values)

To chain several requests, just add more steps in the yaml file starting with the `name` of the step. Use the captured values as explained before.

```
# Create TODO item
POST https://mytodos.com/api/v1/todos/create
{
  "name": "Go shopping",
  "due": "2024-06-09"
}
SAVE todo_id: .id
WAIT 1000
  
# Update the previous TODO
PUT https://mytodos.com/api/v1/todos/{{ .todo_id }}/update
{
  "name": "Go grocery shopping"
} 
```

## Reference

nugget has the following keywords:

- `#`: comments (full line comments supported only)
- `GET, POST, PUT, DELETE, PATCH `: type of request
- `HTTP`: http response code assert
- `HEADER`: add request header
- `SAVE`: capture from response and seve value to variable
- `WAIT`: wait for a certain amount of milliseconds before the next request

And the following pre-defined template variables:

- `{{ .uuid }}`: generates a random UUID

## Author

Isaac Benitez  
June 2024
