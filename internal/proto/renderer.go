package proto

import (
	"fmt"
	"strconv"
	"strings"
)

var respNil = "(nil)"
var emptyList = "(empty list or set)"
var invalidString = "(error) invalid type"

// Define command groups as `map[string]struct{}` for efficient lookups
var simpleStringCommands = map[string]struct{}{
	"AUTH":    {},
	"SELECT":  {},
	"RENAME":  {},
	"RESTORE": {},
	"MSET":    {},
	"SET":     {},
	"PFMERGE": {},
	"FLUSHDB": {},
}

var intCommands = map[string]struct{}{
	"COPY":         {},
	"DEL":          {},
	"EXISTS":       {},
	"EXPIRE":       {},
	"EXPIREAT":     {},
	"EXPIRETIME":   {},
	"PERSIST":      {},
	"PTTL":         {},
	"TTL":          {},
	"TOUCH":        {},
	"HDEL":         {},
	"HEXISTS":      {},
	"HINCRBY":      {},
	"HSET":         {},
	"HSETNX":       {},
	"HSTRLEN":      {},
	"PFADD":        {},
	"PFCOUNT":      {},
	"LPUSH":        {},
	"RPUSH":        {},
	"LLEN":         {},
	"SADD":         {},
	"SREM":         {},
	"SCARD":        {},
	"SETBIT":       {},
	"SETNX":        {},
	"INCR":         {},
	"INCRBY":       {},
	"DECR":         {},
	"DECRBY":       {},
	"APPEND":       {},
	"ZADD":         {},
	"HINCRBYFLOAT": {},
	"BITPOS":       {},
}

var bulkStringCommands = map[string]struct{}{
	"ECHO":         {},
	"PING":         {},
	"DUMP":         {},
	"TYPE":         {},
	"GEODIST":      {},
	"HGET":         {},
	"HINCRBYFLOAT": {},
	"GET":          {},
	"GETEX":        {},
	"GETDEL":       {},
	"GETRANGE":     {},
	"GETSET":       {},
	"INCRBYFLOAT":  {},
	"ZSCORE":       {},
}

var listCommands = map[string]struct{}{
	"HELLO":      {},
	"KEYS":       {},
	"HKEYS":      {},
	"HMGET":      {},
	"HVALS":      {},
	"HRANDFIELD": {},
	"SMEMBERS":   {},
	"SDIFF":      {},
	"SINTER":     {},
	"MGET":       {},
	"BITFIELD":   {},
	"COMMAND":    {},
}

func RenderOutput(cmdName string, cmdVal interface{}, cmdErr error) (interface{}, error) {
	fn := getRender(cmdName)
	if cmdErr != nil {
		return nil, renderError(cmdErr)
	}

	// If we don't have a renderer defined leave it as it is
	if fn == nil {
		return cmdVal, nil
	}

	return fn(cmdVal), nil
}

// getRender retrieves the appropriate callback for the command
func getRender(commandName string) func(value interface{}) interface{} {
	commandUpper := strings.ToUpper(strings.TrimSpace(commandName))

	// Determine the render method based on command group
	if _, exists := simpleStringCommands[commandUpper]; exists {
		return renderSimpleString
	}
	if _, exists := intCommands[commandUpper]; exists {
		return renderInt
	}
	if _, exists := bulkStringCommands[commandUpper]; exists {
		return renderBulkString
	}
	if _, exists := listCommands[commandUpper]; exists {
		return renderList
	}

	return nil
}

func ensureStr(input interface{}) string {
	switch v := input.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func renderError(errorMsg error) error {
	err := ensureStr(errorMsg)

	return fmt.Errorf("(error) %s", err)
}

// Render functions
func renderBulkString(value interface{}) interface{} {
	if value == nil {
		return respNil
	}

	result, ok := value.(int64)
	if ok {
		return renderInt(result)
	}

	return fmt.Sprintf("%v", value)
}

func renderInt(value interface{}) interface{} {
	if value == nil {
		return respNil
	}

	if intValue, ok := value.(int64); ok {
		return fmt.Sprintf("(integer) %d", intValue)
	}

	return invalidString
}

func renderList(value interface{}) interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return invalidString
	}

	var builder strings.Builder
	for i, item := range items {
		// Convert item to string
		if item == nil {
			builder.WriteString(fmt.Sprintf("%d) %v\n", i+1, respNil))
			continue
		}

		strItem := fmt.Sprintf("%v", item)

		// Check if the string is already quoted, if not, add quotes
		if !(strings.HasPrefix(strItem, "\"") && strings.HasSuffix(strItem, "\"")) {
			strItem = fmt.Sprintf("\"%s\"", strItem)
		}

		builder.WriteString(fmt.Sprintf("%d) %s\n", i+1, strItem))
	}
	return builder.String()
}

func renderListOrString(value interface{}) interface{} {
	if items, ok := value.([]interface{}); ok {
		return renderList(items)
	}
	return renderBulkString(value)
}

func renderStringOrInt(value interface{}) interface{} {
	if intValue, ok := value.(int); ok {
		return renderInt(intValue)
	}
	return renderBulkString(value)
}

func renderSimpleString(value interface{}) interface{} {
	if value == nil {
		return respNil
	}

	text := fmt.Sprintf("%v", value)
	return text
}

func renderHashPairs(value interface{}) interface{} {
	items, ok := value.([]interface{})
	if len(items) == 0 {
		return emptyList
	}
	if !ok || len(items)%2 != 0 {
		return "(error) invalid hash pair format"
	}

	var builder strings.Builder
	indexWidth := len(strconv.Itoa(len(items) / 2))
	for i := 0; i < len(items); i += 2 {
		key := fmt.Sprintf("%v", items[i])
		value := fmt.Sprintf("%v", items[i+1])

		// Format the index and key
		indexStr := fmt.Sprintf("%*d) ", indexWidth, i/2+1)
		builder.WriteString(indexStr)
		builder.WriteString(key + "\n")

		// Format the value, ensuring correct indentation
		// and preserving quotes if necessary
		if strings.Contains(value, "\"") {
			value = fmt.Sprintf("%q", value)
		}
		valueStr := strings.Repeat(" ", len(indexStr)) + value
		builder.WriteString(valueStr + "\n")
	}
	return builder.String()
}

func commandHscan(value interface{}) interface{} {
	scanResult, ok := value.([]interface{})
	if !ok || len(scanResult) < 2 {
		return "(error) invalid type or format"
	}

	cursor := fmt.Sprintf("%v", scanResult[0])
	items, ok := scanResult[1].([]interface{})
	if !ok {
		return "(error) invalid scan items format"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(cursor) %s\n", cursor))
	renderedItems := renderHashPairs(items)
	builder.WriteString(fmt.Sprintf("%s", renderedItems))

	return builder.String()
}

// RenderMembers renders a list of set or sorted set members
func renderMembers(value interface{}) interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return invalidString
	}

	var builder strings.Builder
	indexWidth := len(strconv.Itoa(len(items)))
	for i, item := range items {
		member := fmt.Sprintf("%v", item)
		indexStr := fmt.Sprintf("%*d) ", indexWidth, i+1)
		builder.WriteString(indexStr)
		builder.WriteString(member + "\n")
	}

	return builder.String()
}
