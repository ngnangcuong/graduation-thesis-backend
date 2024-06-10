local http = require "resty.http"
local utils = require "kong.tools.utils"

local TokenHandler = {
    VERSION = "1.0",
    PRIORITY = 1000,
}


local function introspect_access_token(conf, access_token)
  local httpc = http:new()
  -- step 1: validate the token
  local res, err = httpc:request_uri(conf.introspection_endpoint, {
      method = "POST",
      ssl_verify = false,
      headers = {
          ["Content-Type"] = "application/x-www-form-urlencoded",
          ["Authorization"] = "Bearer " .. access_token }
  })

  if not res then
      kong.log.err("failed to call introspection endpoint: ", err)
      return kong.response.exit(500)
  end
  if res.status ~= 200 then
      kong.log.err("introspection endpoint responded with status: ", res.status, res.body)
      return kong.response.exit(res.status)
  end

  local body = res.body

  return body -- all is well
end

function TokenHandler:access(conf)
  local access_token = kong.request.get_headers()[conf.token_header]
  if not access_token then
      kong.response.exit(401)  --unauthorized
  end
  -- replace Bearer prefix
  access_token = access_token:sub(8,-1) -- drop "Bearer "

  local user_id = introspect_access_token(conf, access_token)
  kong.service.request.set_header("X-User-ID", user_id)

end


return TokenHandler