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

Pass the `-q` flag to print just the response body.

```bash
% nugget requests.n -q
```

## Minimal example

This would be a simple nugget file with a get request:

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

## Save values

Use the `SAVE` keyword to save one or more values from the response:

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

To save values, use the path of the value you want to save from the response. For example, `.address.country` if you need to save the country in the following json:

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

## Use saved values

You can use the saved values adding the variable name in a "template" like fashion: `{{ .variable-name }}`.

The saved values can be use in the following areas:

- The body json
- The url
- The header

## Chain several requests (use saved values)

To chain several requests, just add more requests to the same nugget file. Use the saved values as explained before.

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
# Stop until ENTER is pressed
WAIT -1

# Update the previous TODO again
PUT https://mytodos.com/api/v1/todos/{{ .todo_id }}/update
{
  "name": "Go shopping"
} 
```

## Reference

nugget has the following keywords:

- `#`: comments (full line comments supported only)
- `GET, POST, PUT, DELETE, PATCH `: type of request
- `HTTP`: http response code assert
- `HEADER`: add request header
- `SAVE`: save a value from the response to a variable
- `WAIT`: wait for a certain amount of milliseconds before the next request; use `-1` to continue after `ENTER` is pressed

And the following pre-defined template variables:

- `{{ .uuid }}`: generates a random UUID

## Author

Isaac Benitez  
June 2024
