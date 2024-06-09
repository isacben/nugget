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

## Capture values

Use the capture keyword to capture a list of values from the response:

```yaml
steps:                                                                                                                             - name: Create TODO item
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

To capture values, use the keys in the response concatenated with the '.' (in a jq fashion). For example, `.address.country`. 

```yaml
steps:                                                                                                                             - name: Create TODO item
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

## Author

Isaac Benitez  
June 2024