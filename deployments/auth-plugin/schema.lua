local typedefs = require "kong.db.schema.typedefs"


return {
  name = "custom-auth",
  fields = {
    { protocols = typedefs.protocols_http },
    { consumer = typedefs.no_consumer },
    { config = {
        type = "record",
        fields = {
          { introspection_endpoint = typedefs.url({ required = true }) },
          { token_header = typedefs.header_name { default = "Authorization", required = true }, }
    }, }, },
  },
}
