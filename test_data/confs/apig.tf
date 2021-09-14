resource "aws_apigatewayv2_route" "{ep_name}" {
  api_id = "{api_id}"
  route_key = "{http_method} {route_path}"
  target = "integs/${aws_apigatewayv2_integration.{ep_name}.id}"
}

