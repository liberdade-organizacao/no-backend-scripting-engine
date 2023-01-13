function split_string(inlet, sep)
 if sep == nil then
  sep = "%s"
 end
 local t={}
 for s in string.gmatch(inlet, "([^"..sep.."]+)") do
  table.insert(t, s)
 end
 return t
end

function main(param)
 local key_value_pairs = split_string(param, "&")
 local outlet = "not found!"
 for _, key_value_pair in pairs(key_value_pairs) do
  local stuff = split_string(key_value_pair, "=")
  local key = stuff[1]
  local value = stuff[2]
  if key == "name" then
   outlet = value
  end
 end
 return outlet
end

