local outOfStock = {}

for i, key in ipairs(KEYS) do
    local currentStock = tonumber(redis.call("GET", key) or "0")
    local delta = tonumber(ARGV[i])
    if delta < 0 and currentStock < -delta then
        table.insert(outOfStock, key)
    end
end

if #outOfStock > 0 then 
    return outOfStock
end

for i, key in ipairs(KEYS) do
    local delta = tonumber(ARGV[i])
    if delta < 0 then
        redis.call("DECRBY", key, -delta)
    else
        redis.call("INCRBY", key, delta)
    end
end

return outOfStock