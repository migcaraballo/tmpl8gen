# tmpl8gen
make templates for anything and everything.

## what does this do?

In simplest terms, take a mapping file in the form of json, copies a directory of files, and applies the mapping to the new copy of the directory, leaving the "template source" intact. The mappings are applied to all files in the template directory, including subfolders.

## sample usage
```shell script
Usage of ./tmpl8gen:
  -bc
    	bypass confirmation prompts
  -map_path string
    	path to template map file
  -out_dir string
    	path for generated output
  -tmp_path string
    	path to template directory
```

## how to build & run
After executing this, you will have a binary for your OS in the project root
```shell script
project_root_on_disk> ./buildrun.sh
```

## flag info
flag name | description
--------- | -----------
bc | bypasses all confirmation prompts (true/false, or just -bc)
map_path | the json mapping file path (can be relative or absolute)
out_dir | the output of your template directory (can be relative or absolute)
tmp_path | the path to the template source directory (can be relative or absolute) 

## how to create your templates?
Templates can be created for anything. Anything you want to templatize, just create a mapping in the json file, then in the template file use the field from the json mapping surrounded by curly brackets.

### example mapping
```json
{
  "api_id": "ps-1231",
  "ep_name": "petsearch",
  "route_path": "/pet/search",
  "http_method": "GET",
  "fren_name": "Pet Search",
  "app_code": "PTSRCH",
  "role_code": "SERVICE"
}
```
### example template
```hcl-terraform
resource "aws_apigatewayv2_route" "{ep_name}" {
  api_id = "{api_id}"
  route_key = "{http_method} {route_path}"
  target = "integs/${aws_apigatewayv2_integration.{ep_name}.id}"
}
```

### example output
```hcl-terraform
resource "aws_apigatewayv2_route" "petsearch" {
  api_id = "ps-1231"
  route_key = "GET /pet/search"
  target = "integs/${aws_apigatewayv2_integration.petsearch.id}"
}
```

### run the example
```shell script
# mac/linxu/unix
./tmpl8gen -map_path=test_data/template_maps/lambda_tmp_map.json -tmp_path=test_data/confs/ -out_dir=output/app/confs/
```
```shell script
# windows
C:\> path-to-project-root\tmpl8gen.exe -map_path=test_data/template_maps/lambda_tmp_map.json -tmp_path=test_data/confs/ -out_dir=output/app/confs/
```