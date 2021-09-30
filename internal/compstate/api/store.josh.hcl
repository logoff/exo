interface "store" {
  method "set-state" {
    input "component-id" "string" {}
    input "type" "string" {}
    input "content" "string" {}
    input "tags" "map[string]string" {}
    input "timestamp" "string" {}
    
    output "version" "int" {}
  }
  
  method "get-state" {
    input "component-id" "string" {}
    
    output "state" "*State" {}
  }
  
  // TODO: describe-components?
  // TODO: garbage collection?
}

struct "state" {
  field "component-id" "string" {}
  field "version" "int" {}
  field "type" "string" {}
  field "content" "string" {}
  field "tags" "map[string]string" {}
  field "timestamp" "string" {}
}
