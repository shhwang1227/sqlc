# a support in version sqlc: A SQL Compiler

sqlc generates **type-safe code** from SQL. Here's how it works:

1. You write queries in SQL.
1. You run sqlc to generate code with type-safe interfaces to those queries.
1. You write application code that calls the generated code.
1. support in syntax
Check out [an interactive example](https://github.com/xiazemin/sqlc_study) to see it in action.


#安装
 go get -u github.com/xiazemin/sqlc
 go get -u github.com/xiazemin/sqlc/cmd/sqlc
#使用
实例：https://github.com/xiazemin/sqlc_study

## Sponsors

sqlc development is funded by our [generous
sponsors](https://github.com/sponsors/xiazemin), including the following
companies:

If you use sqlc at your company, please consider [becoming a
sponsor](https://github.com/sponsors/xiazemin) today.

Sponsors receive priority support via the sqlc Slack organization.

支持驼峰格式
支持生成mock代码

 //go:generate  mockgen -source=./querier.go  -destination=./mock/querier.go

 go generate ./... 

 解决聚合函数返回值是interface{}
 必须用下面函数解析的问题
 ```
 func ParseInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, nil
	}
	raw, ok := v.([]uint8)
	if !ok {
		return 0, fmt.Errorf("type assert failed")
	}
	return strconv.ParseInt(string(raw), 10, 64)
}
```